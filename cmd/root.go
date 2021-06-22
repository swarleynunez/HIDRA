package cmd

import (
	"context"
	"github.com/spf13/cobra"
)

var (
	// Main context
	ctx = context.Background()

	// Root CLI command
	rootCmd = &cobra.Command{
		Use:  "hidra",
		Long: "HIDRA distributed container orchestrator",
	}
)

func init() {

	// CLI init configuration
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	// CLI available commands
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(registerCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(versionCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
