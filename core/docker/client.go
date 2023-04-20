package docker

import (
	"context"
	"github.com/docker/docker/client"
	"github.com/swarleynunez/hidra/core/utils"
)

func Connect(ctx context.Context) (docc *client.Client) {

	docc, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	utils.CheckError(err, utils.FatalMode)

	// Check connection
	_, err = docc.Ping(ctx)
	utils.CheckError(err, utils.FatalMode)

	return
}
