package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tunajam/gh-packs/internal/git"
	"github.com/tunajam/gh-packs/internal/manifest"
)

var ValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Lint packs.yaml and verify lockfile integrity",
	RunE: func(cmd *cobra.Command, args []string) error {
		errors := 0

		// Check manifest
		m, err := manifest.Load(manifest.ManifestFile)
		if err != nil {
			return fmt.Errorf("manifest error: %w", err)
		}

		if m.Name == "" {
			fmt.Println("✗ manifest: missing 'name'")
			errors++
		}
		if m.Version == "" {
			fmt.Println("✗ manifest: missing 'version'")
			errors++
		}

		for name, dep := range m.Dependencies {
			if dep.Version == "" {
				fmt.Printf("✗ dependency %s: missing 'version'\n", name)
				errors++
			}
			if dep.Source == "" {
				fmt.Printf("✗ dependency %s: missing 'source'\n", name)
				errors++
			}
		}

		// Check lockfile integrity
		lockfile, err := manifest.LoadLockfile(manifest.LockFile)
		if err != nil {
			if os.IsNotExist(err) && len(m.Dependencies) > 0 {
				fmt.Println("✗ lockfile missing but dependencies declared (run 'gh packs install')")
				errors++
			}
		} else {
			for key, pkg := range lockfile.Packages {
				// Extract repo name
				parts := strings.Split(key, "/")
				nameWithVer := parts[len(parts)-1]
				repoName := strings.Split(nameWithVer, "@")[0]

				packDir := filepath.Join(".context", "packs", repoName)
				if _, err := os.Stat(packDir); os.IsNotExist(err) {
					fmt.Printf("✗ %s: not installed\n", key)
					errors++
					continue
				}

				hash, err := git.HashDir(packDir)
				if err != nil {
					fmt.Printf("✗ %s: hash error: %v\n", key, err)
					errors++
					continue
				}

				if hash != pkg.ContentHash {
					fmt.Printf("✗ %s: integrity mismatch\n", key)
					errors++
				} else {
					fmt.Printf("✓ %s\n", key)
				}
			}
		}

		if errors > 0 {
			return fmt.Errorf("%d validation errors", errors)
		}
		fmt.Println("\n✓ All checks passed")
		return nil
	},
}
