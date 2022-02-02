package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/swarleynunez/superfog/core/managers"
)

const showShortMsg = "Show active cluster applications and containers"

var showCmd = &cobra.Command{
	Use:                   "show",
	Short:                 showShortMsg,
	Long:                  title + "\n\n" + "Info:\n  " + showShortMsg,
	DisableFlagsInUseLine: true,
	Args:                  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize and configure node
		managers.InitNode(ctx, false)

		/*// Get flags
		owned, err := cmd.Flags().GetBool("owned")
		utils.CheckError(err, utils.FatalMode)

		// Filter active cluster applications
		apps := managers.GetActiveApplications()
		if owned {
			for appid, app := range apps {
				if app.Owner != managers.GetFromAccount() {
					delete(apps, appid)
				}
			}
		}*/

		// Print cluster applications
		apps := managers.GetActiveApplications()
		if len(apps) == 0 {
			fmt.Println("--> No cluster applications")
			return
		}
		for appid, app := range apps {
			fmt.Println("--> APPID:", appid)
			fmt.Println("    OWNER:", app.Owner)
			fmt.Println("    REGISTERED:", app.RegisteredAt)

			// Print application's containers
			ctrs := managers.GetApplicationContainersData(appid)
			for rcid, ctr := range ctrs {
				// Get container host
				insts := managers.GetContainerInstances(rcid)

				fmt.Println("\t\tRCID:", rcid)
				fmt.Println("\t\tHOST:", insts[len(insts)-1].Host)
				fmt.Println("\t\tREGISTERED:", ctr.RegisteredAt)
			}
		}
	},
}
