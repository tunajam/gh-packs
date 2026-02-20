package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tunajam/gh-packs/internal/manifest"
)

var RemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a dependency",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		m, err := manifest.Load(manifest.ManifestFile)
		if err != nil {
			return err
		}

		// Find matching dependency
		found := ""
		for depName := range m.Dependencies {
			if depName == name || strings.HasSuffix(depName, "/"+name) {
				found = depName
				break
			}
		}
		if found == "" {
			return fmt.Errorf("dependency %q not found", name)
		}

		// Remove from manifest
		delete(m.Dependencies, found)
		if err := m.Save(manifest.ManifestFile); err != nil {
			return err
		}

		// Remove from lockfile
		lockfile, err := manifest.LoadLockfile(manifest.LockFile)
		if err == nil {
			for key := range lockfile.Packages {
				if strings.HasPrefix(key, found+"@") {
					delete(lockfile.Packages, key)
				}
			}
			lockfile.Save(manifest.LockFile)
		}

		// Remove installed files
		// Extract repo name from found (e.g., "@org/repo" -> "repo")
		parts := strings.Split(found, "/")
		repoName := parts[len(parts)-1]
		packDir := filepath.Join(".context", "packs", repoName)
		os.RemoveAll(packDir)

		fmt.Printf("Removed %s\n", found)
		return nil
	},
}
