package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tunajam/gh-packs/internal/git"
	"github.com/tunajam/gh-packs/internal/manifest"
)

var addVersion string
var addPath string
var addReason string

var AddCmd = &cobra.Command{
	Use:   "add <source>",
	Short: "Add a dependency from a GitHub repo",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		source := args[0]
		owner, repo, err := git.ParseSource(source)
		if err != nil {
			return err
		}

		fmt.Printf("Scanning github.com/%s/%s via gh...\n", owner, repo)

		// Load or create manifest
		m, err := manifest.Load(manifest.ManifestFile)
		if err != nil {
			return fmt.Errorf("load manifest: %w (run 'gh packs init' first)", err)
		}

		// List tags
		tags, err := git.ListTags(owner, repo)
		if err != nil {
			return fmt.Errorf("listing tags: %w", err)
		}

		if len(tags) == 0 {
			return fmt.Errorf("no tags found for %s/%s", owner, repo)
		}

		fmt.Printf("Tags: %s (latest)", tags[0].Name)
		for i := 1; i < len(tags) && i < 4; i++ {
			fmt.Printf(", %s", tags[i].Name)
		}
		fmt.Println()

		// Resolve version
		vRange := addVersion
		if vRange == "" {
			// Default to caret of latest
			vRange = "^" + strings.TrimPrefix(tags[0].Name, "v")
		}

		tag, err := git.ResolveVersion(tags, vRange)
		if err != nil {
			return fmt.Errorf("version resolution: %w", err)
		}
		fmt.Printf("Resolving %s → %s\n\n", vRange, tag.Name)

		// Determine context path
		contextPath := addPath
		if contextPath == "" {
			contextPath = ".context"
		}

		// Sparse checkout
		packName := fmt.Sprintf("@%s/%s", owner, repo)
		destDir := filepath.Join(".context", "packs", repo)
		os.MkdirAll(destDir, 0755)

		fmt.Printf("Sparse checkout: %s/ at %s (commit %s)\n\n", contextPath, tag.Name, tag.Commit.SHA[:7])

		files, err := git.FetchContextDir(owner, repo, tag.Name, contextPath, destDir)
		if err != nil {
			return fmt.Errorf("fetching context: %w", err)
		}

		// Compute hash
		hash, err := git.HashDir(destDir)
		if err != nil {
			return fmt.Errorf("hashing: %w", err)
		}

		// Calculate size
		var totalSize int64
		for _, f := range files {
			if info, err := os.Stat(f); err == nil {
				totalSize += info.Size()
			}
		}

		// Update manifest
		if m.Dependencies == nil {
			m.Dependencies = make(map[string]manifest.Dependency)
		}
		dep := manifest.Dependency{
			Version: vRange,
			Source:  fmt.Sprintf("github.com/%s/%s", owner, repo),
			Reason:  addReason,
		}
		if contextPath != ".context" {
			dep.Path = contextPath
		}
		m.Dependencies[packName] = dep
		if err := m.Save(manifest.ManifestFile); err != nil {
			return err
		}

		// Update lockfile
		lockfile, err := manifest.LoadLockfile(manifest.LockFile)
		if err != nil {
			lockfile = manifest.NewLockfile()
		}
		lockKey := fmt.Sprintf("%s@%s", packName, strings.TrimPrefix(tag.Name, "v"))
		lockfile.Packages[lockKey] = manifest.LockedPack{
			Source:      fmt.Sprintf("github.com/%s/%s", owner, repo),
			Ref:         tag.Name,
			Commit:      tag.Commit.SHA,
			ContentHash: hash,
		}
		lockfile.ResolvedAt = time.Now().UTC().Format(time.RFC3339)
		if err := lockfile.Save(manifest.LockFile); err != nil {
			return err
		}

		fmt.Printf("  + %s@%s       (%.1f KB)\n\n", packName, strings.TrimPrefix(tag.Name, "v"), float64(totalSize)/1024)
		fmt.Println("Updated packs.yaml")
		fmt.Println("Updated packs.lock")
		fmt.Println("Installed to .context/packs/")

		return nil
	},
}

func init() {
	AddCmd.Flags().StringVar(&addVersion, "version", "", "semver range (default: ^latest)")
	AddCmd.Flags().StringVar(&addPath, "path", "", "context path in source repo (default: .context)")
	AddCmd.Flags().StringVar(&addReason, "reason", "", "why this dependency is needed")
}
