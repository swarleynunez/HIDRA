package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/swarleynunez/hidra/core/managers"
)

const appRemoveShortMsg = "Remove an application from the cluster"

var appRemoveCmd = &cobra.Command{
	Use:                   "remove APPID",
	Short:                 appRemoveShortMsg,
	Long:                  title + "\n\n" + "Info:\n  " + appRemoveShortMsg,
	DisableFlagsInUseLine: true,
	Args:                  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize and configure node
		managers.InitNode(ctx, false)

		// Get and format args
		/*appid, err := strconv.ParseUint(args[0], 10, 64)
		utils.CheckError(err, utils.FatalMode)

		err = managers.RemoveDCRApplication(ctx, appid)
		utils.CheckError(err, utils.FatalMode)*/

		fmt.Println("--> Application removed from the cluster")
	},
}
