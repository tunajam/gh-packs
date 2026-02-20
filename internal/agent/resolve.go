package agent

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tunajam/gh-packs/internal/manifest"
)

// Resolve spawns a sub-agent via claude --print to synthesize a task briefing.
func Resolve(m *manifest.Manifest, packsDir string, task string, w io.Writer) error {
	// Load system prompt
	systemPrompt := DefaultResolvePrompt
	promptPath := filepath.Join(".packs", "resolve-prompt.md")
	if data, err := os.ReadFile(promptPath); err == nil {
		systemPrompt = string(data)
	}

	// Append resolve config
	if m.Resolve.SystemPromptAppend != "" {
		systemPrompt += "\n" + m.Resolve.SystemPromptAppend
	}

	// Read manifest
	manifestData, err := os.ReadFile(manifest.ManifestFile)
	if err != nil {
		return fmt.Errorf("reading manifest: %w", err)
	}

	// Assemble pack contents
	var packContent strings.Builder
	entries, err := os.ReadDir(packsDir)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading packs dir: %w", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		packContent.WriteString(fmt.Sprintf("\n### %s\n\n", entry.Name()))
		packDir := filepath.Join(packsDir, entry.Name())
		filepath.WalkDir(packDir, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			rel, _ := filepath.Rel(packDir, path)
			data, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			packContent.WriteString(fmt.Sprintf("#### %s\n\n```\n%s\n```\n\n", rel, string(data)))
			return nil
		})
	}

	// Build prompt
	prompt := fmt.Sprintf("## Dependency Manifest\n\n```yaml\n%s\n```\n\n## Installed Packs\n%s\n## Task\n\n%s",
		string(manifestData), packContent.String(), task)

	// Spawn claude --print
	cmd := exec.Command("claude", "--print", "--system-prompt", systemPrompt, prompt)
	cmd.Stdout = w
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
