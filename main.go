package main

import (
	"os"

	"github.com/tunajam/gh-packs/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
