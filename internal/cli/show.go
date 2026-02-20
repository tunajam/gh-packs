package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var ShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Dump a single installed pack to stdout",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		packDir := filepath.Join(".context", "packs", name)

		if _, err := os.Stat(packDir); os.IsNotExist(err) {
			return fmt.Errorf("pack %q not installed", name)
		}

		return filepath.WalkDir(packDir, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			rel, _ := filepath.Rel(packDir, path)
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			fmt.Printf("### %s\n\n%s\n\n", rel, string(data))
			return nil
		})
	},
}
