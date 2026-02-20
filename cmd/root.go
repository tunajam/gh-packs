package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tunajam/gh-packs/internal/cli"
)

var rootCmd = &cobra.Command{
	Use:   "gh-packs",
	Short: "The package manager for AI context",
	Long:  "packs — versioned context packages across repositories for AI coding assistants",
}

func init() {
	rootCmd.AddCommand(cli.InitCmd)
	rootCmd.AddCommand(cli.AddCmd)
	rootCmd.AddCommand(cli.InstallCmd)
	rootCmd.AddCommand(cli.ResolveCmd)
	rootCmd.AddCommand(cli.ShowCmd)
	rootCmd.AddCommand(cli.GraphCmd)
	rootCmd.AddCommand(cli.RemoveCmd)
	rootCmd.AddCommand(cli.ValidateCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
