package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/swarleynunez/hidra/core/managers"
	"github.com/swarleynunez/hidra/core/types"
	"github.com/swarleynunez/hidra/core/utils"
	"github.com/swarleynunez/hidra/inputs"
)

const appDeployShortMsg = "Deploy a new application on the cluster"

var appDeployCmd = &cobra.Command{
	Use:                   "deploy [OPTIONS]",
	Short:                 appDeployShortMsg,
	Long:                  title + "\n\n" + "Info:\n  " + appDeployShortMsg,
	DisableFlagsInUseLine: true,
	Args:                  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize and configure node
		managers.InitNode(ctx, false)

		// Get flags
		autodeploy, err := cmd.Flags().GetBool("autodeploy")
		utils.CheckError(err, utils.FatalMode)

		//fmt.Println("--> Starting at", time.Now().UnixMilli())

		// TODO. SDN ONOS plugin: check if the new application (VS) already exists
		err = managers.RegisterApplication(ctx, &inputs.AppInfo, []types.ContainerInfo{inputs.CtrInfo}, autodeploy)
		utils.CheckError(err, utils.FatalMode)

		fmt.Println("--> Application deployed on the cluster")
	},
}
