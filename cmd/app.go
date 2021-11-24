package cmd

import (
	"github.com/spf13/cobra"
)

const appShortMsg = "Manage applications"

var appCmd = &cobra.Command{
	Use:                   "application",
	Aliases:               []string{"app"},
	Short:                 appShortMsg,
	Long:                  title + "\n\n" + "Info:\n  " + appShortMsg,
	DisableFlagsInUseLine: true,
}
