package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tunajam/gh-packs/internal/agent"
	"github.com/tunajam/gh-packs/internal/git"
	"github.com/tunajam/gh-packs/internal/manifest"
)

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Scaffold packs.yaml, .context/, and .packs/",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if manifest already exists
		if _, err := os.Stat(manifest.ManifestFile); err == nil {
			return fmt.Errorf("packs.yaml already exists")
		}

		// Detect repo info
		name := "@org/my-service"
		owner, repo, err := git.DetectRemote()
		if err == nil {
			name = fmt.Sprintf("@%s/%s", owner, repo)
		}

		// Create manifest
		m := &manifest.Manifest{
			Name:         name,
			Version:      "1.0.0",
			Description:  "Service description",
			Pack:         manifest.PackInfo{Type: "service"},
			Dependencies: make(map[string]manifest.Dependency),
			Resolve:      manifest.ResolveConfig{MaxDepth: 2},
		}
		if err := m.Save(manifest.ManifestFile); err != nil {
			return err
		}

		// Create .context/
		contextDir := ".context"
		os.MkdirAll(contextDir, 0755)

		contextMD := filepath.Join(contextDir, "CONTEXT.md")
		if _, err := os.Stat(contextMD); os.IsNotExist(err) {
			content := fmt.Sprintf("# %s\n\n## Overview\n\nDescribe your service here.\n\n## Architecture\n\n## API\n\n## Conventions\n", name)
			os.WriteFile(contextMD, []byte(content), 0644)
		}

		// Create .packs/
		os.MkdirAll(".packs", 0755)
		promptPath := filepath.Join(".packs", "resolve-prompt.md")
		if _, err := os.Stat(promptPath); os.IsNotExist(err) {
			os.WriteFile(promptPath, []byte(agent.DefaultResolvePrompt), 0644)
		}

		fmt.Printf("Detected: github.com/%s/%s\n\n", owner, repo)
		fmt.Println("Created:")
		fmt.Println("  packs.yaml               ← manifest (edit to add dependencies)")
		fmt.Println("  .context/CONTEXT.md      ← your repo's context template (fill this in)")
		fmt.Println("  .packs/resolve-prompt.md ← sub-agent system prompt (customizable)")

		return nil
	},
}
