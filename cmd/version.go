package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

const versionShortMsg = "Show version information"

var versionCmd = &cobra.Command{
	Use:                   "version",
	Short:                 versionShortMsg,
	Long:                  title + "\n\n" + "Info:\n  " + versionShortMsg,
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("--> HIDRA 1.0.0")
	},
}
