package docker

import (
	"github.com/docker/docker/client"
	"github.com/swarleynunez/superfog/core/utils"
)

func Connect() (docc *client.Client) {

	docc, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	utils.CheckError(err, utils.WarningMode)

	return
}
