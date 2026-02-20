package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tunajam/gh-packs/internal/manifest"
)

var GraphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Print the dependency tree",
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := manifest.Load(manifest.ManifestFile)
		if err != nil {
			return fmt.Errorf("load manifest: %w", err)
		}

		lockfile, _ := manifest.LoadLockfile(manifest.LockFile)

		fmt.Printf("%s (you are here)\n", m.Name)
		deps := make([]string, 0, len(m.Dependencies))
		for name := range m.Dependencies {
			deps = append(deps, name)
		}

		for i, name := range deps {
			prefix := "├── "
			if i == len(deps)-1 {
				prefix = "└── "
			}

			version := ""
			if lockfile != nil {
				for key := range lockfile.Packages {
					if strings.HasPrefix(key, name+"@") {
						version = strings.TrimPrefix(key, name+"@")
						break
					}
				}
			}

			if version != "" {
				fmt.Printf("  %s%s@%s\n", prefix, name, version)
			} else {
				fmt.Printf("  %s%s\n", prefix, name)
			}
		}

		return nil
	},
}
