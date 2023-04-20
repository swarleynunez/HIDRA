package cmd

import (
	"context"
	"github.com/spf13/cobra"
)

var (
	// Main context
	ctx = context.Background()

	// Main CLI title
	title = `------------------------------------------------
--- HIDRA distributed container orchestrator ---
------------------------------------------------`

	// Root CLI command
	rootCmd = &cobra.Command{
		Use:  "hidra",
		Long: title,
	}
)

func init() {

	// CLI init configuration
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	// CLI available commands
	rootCmd.AddCommand(
		deployCmd,
		registerCmd,
		runCmd,
		appCmd,
		showCmd,
		versionCmd)

	// Subcommands
	appCmd.AddCommand(appDeployCmd)
	appCmd.AddCommand(appRemoveCmd)

	// Flags
	appDeployCmd.Flags().BoolP("autodeploy", "a", false, "deploy application in autodeploy mode")
	//showCmd.Flags().BoolP("owned", "o", false, "show cluster applications owned by this node")
}

func Execute() error {
	return rootCmd.Execute()
}
