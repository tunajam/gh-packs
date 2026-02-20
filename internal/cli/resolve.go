package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tunajam/gh-packs/internal/agent"
	"github.com/tunajam/gh-packs/internal/manifest"
)

var resolveDryRun bool

var ResolveCmd = &cobra.Command{
	Use:   `resolve "<task>"`,
	Short: "Spawn sub-agent to synthesize task-specific briefing",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		task := args[0]

		m, err := manifest.Load(manifest.ManifestFile)
		if err != nil {
			return fmt.Errorf("load manifest: %w", err)
		}

		packsDir := filepath.Join(".context", "packs")
		if _, err := os.Stat(packsDir); os.IsNotExist(err) {
			return fmt.Errorf("no packs installed (run 'gh packs install' first)")
		}

		if resolveDryRun {
			fmt.Println("Would load packs from:")
			entries, _ := os.ReadDir(packsDir)
			for _, e := range entries {
				if e.IsDir() {
					fmt.Printf("  - %s\n", e.Name())
				}
			}
			fmt.Printf("\nTask: %s\n", task)
			return nil
		}

		return agent.Resolve(m, packsDir, task, os.Stdout)
	},
}

func init() {
	ResolveCmd.Flags().BoolVar(&resolveDryRun, "dry-run", false, "show which packs would load")
}
