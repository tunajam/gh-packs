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

var InstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install all dependencies from lockfile",
	RunE: func(cmd *cobra.Command, args []string) error {
		lockfile, err := manifest.LoadLockfile(manifest.LockFile)
		if err != nil {
			return fmt.Errorf("no lockfile found (run 'gh packs add' first): %w", err)
		}

		// Also load manifest for path info
		m, _ := manifest.Load(manifest.ManifestFile)

		fmt.Println("Installing from packs.lock...")
		start := time.Now()
		count := 0

		for key, pkg := range lockfile.Packages {
			owner, repo, err := git.ParseSource(pkg.Source)
			if err != nil {
				return fmt.Errorf("invalid source for %s: %w", key, err)
			}

			// Extract pack name from key (e.g., "@org/repo@1.0.0" -> "repo")
			packName := key
			if idx := strings.LastIndex(key, "@"); idx > 0 {
				packName = key[:idx]
			}

			destDir := filepath.Join(".context", "packs", repo)
			os.MkdirAll(destDir, 0755)

			// Determine context path
			contextPath := ".context"
			if m != nil {
				for _, dep := range m.Dependencies {
					if dep.Source == pkg.Source {
						contextPath = dep.ContextPath()
						break
					}
				}
			}

			_, err = git.FetchContextDir(owner, repo, pkg.Ref, contextPath, destDir)
			if err != nil {
				return fmt.Errorf("installing %s: %w", key, err)
			}

			// Verify integrity
			hash, err := git.HashDir(destDir)
			if err != nil {
				return fmt.Errorf("hashing %s: %w", key, err)
			}

			if hash != pkg.ContentHash {
				// Remove failed install
				os.RemoveAll(destDir)
				return fmt.Errorf("integrity check failed for %s: expected %s, got %s", key, pkg.ContentHash, hash)
			}

			_ = packName
			fmt.Printf("  %s  ✓ verified\n", key)
			count++
		}

		fmt.Printf("\n%d packs installed in %.1fs\n", count, time.Since(start).Seconds())
		return nil
	},
}
