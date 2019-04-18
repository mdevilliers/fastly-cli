package main

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "fastly-cli",
}

func main() {
	rootCmd.AddCommand(launchCommand)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
