package managers

import (
	"context"
	"errors"
	"fmt"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/swarleynunez/hidra/core/types"
	"github.com/swarleynunez/hidra/core/utils"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

const (
	cnameTemplate = "hidra.io_rcid-"
)

var (
	errContainerNotFound = errors.New("container not found")
)

// //////////
// Images //
// //////////
func existImageLocally(ctx context.Context, imgTag string) bool {

	// Check and format tag
	imgTag, err := utils.FormatImageTag(imgTag)
	utils.CheckError(err, utils.WarningMode)

	// Get all local images
	images, err := _docc.ImageList(ctx, dockertypes.ImageListOptions{All: true})
	utils.CheckError(err, utils.WarningMode)

	// Search image by tag
	for i := range images {
		for j := range images[i].RepoTags {
			if images[i].RepoTags[j] == imgTag {
				return true
			}
		}
	}

	return false
}

func pullImage(ctx context.Context, imgTag string) {

	// Debug
	fmt.Print("[", time.Now().UnixMilli(), "] ", "Downloading '"+imgTag+"' image...\n")

	out, err := _docc.ImagePull(ctx, imgTag, dockertypes.ImagePullOptions{})
	utils.CheckError(err, utils.WarningMode)
	_, err = io.Copy(ioutil.Discard, out) // Discard output to /dev/null
	utils.CheckError(err, utils.WarningMode)
}

func pruneImages(ctx context.Context) {

	_, err := _docc.ImagesPrune(ctx, filters.Args{})
	utils.CheckError(err, utils.WarningMode)
}

// //////////////
// Containers //
// //////////////
func createDockerContainer(ctx context.Context, cinfo *types.ContainerInfo, cname string) {

	// Check and format image tag
	imgTag, err := utils.FormatImageTag(cinfo.ImageTag)
	utils.CheckError(err, utils.WarningMode)

	if !existImageLocally(ctx, imgTag) {
		pullImage(ctx, imgTag)
	}

	ports := checkNodePorts(ctx, cinfo.Ports)

	// Set container configs
	ctrConfig := &container.Config{
		Env:   cinfo.Envs,
		Image: imgTag,
	}
	hostConfig := &container.HostConfig{
		Binds:        cinfo.Volumes,
		PortBindings: ports,
		Resources: container.Resources{
			Memory: int64(cinfo.MemLimit),
			//NanoCPUs: int64(cinfo.CPULimit),
		},
	}
	netConfig := &network.NetworkingConfig{}

	_, err = _docc.ContainerCreate(ctx, ctrConfig, hostConfig, netConfig, nil, cname)
	utils.CheckError(err, utils.WarningMode)
}

func startDockerContainer(ctx context.Context, cname string) {

	err := _docc.ContainerStart(ctx, cname, dockertypes.ContainerStartOptions{})
	utils.CheckError(err, utils.WarningMode)
}

func restartDockerContainer(ctx context.Context, cname string) {

	// Stop and start container
	err := _docc.ContainerRestart(ctx, cname, container.StopOptions{})
	utils.CheckError(err, utils.WarningMode)
}

func renameDockerContainer(ctx context.Context, cname, new string) {

	err := _docc.ContainerRename(ctx, cname, new)
	utils.CheckError(err, utils.WarningMode)
}

func stopDockerContainer(ctx context.Context, cname string) {

	// SIGTERM instead of SIGKILL
	err := _docc.ContainerStop(ctx, cname, container.StopOptions{})
	utils.CheckError(err, utils.WarningMode)
}

func removeDockerContainer(ctx context.Context, cname string) {

	err := _docc.ContainerRemove(ctx, cname, dockertypes.ContainerRemoveOptions{})
	utils.CheckError(err, utils.WarningMode)
}

/*func BackupContainer() {

	// TODO. Backup tasks. Improve flow
	// commit --> create an image from a container (snapshot preserving rw)
	// save, load --> compress and decompress images (tar or stdin/stdout)
	// volumes --> manual backup or using --volumes-from (temporal container)
}*/

// all: only running containers (false) or all containers (true)
func SearchDockerContainers(ctx context.Context, key, value string, all bool) []dockertypes.Container {

	filter := filters.Args{}
	if key != "" && value != "" {
		filter = filters.NewArgs(filters.KeyValuePair{Key: key, Value: value})
	}

	ctrs, err := _docc.ContainerList(ctx, dockertypes.ContainerListOptions{Size: true, All: all, Filters: filter})
	utils.CheckError(err, utils.WarningMode)

	if len(ctrs) > 0 {
		return ctrs
	} else {
		return nil
	}
}

// ///////////
// Volumes //
// ///////////
func pruneVolumes(ctx context.Context) {

	_, err := _docc.VolumesPrune(ctx, filters.Args{})
	utils.CheckError(err, utils.WarningMode)
}

// ///////////
// Helpers //
// ///////////
// Format cname from a rcid
func GetContainerName(rcid uint64) string {
	return cnameTemplate + strconv.FormatUint(rcid, 10)
}

// Extract the registry container ID (rcid) from a cname
func getRegContainerId(cname string) uint64 {

	// Subtract template substring
	s := strings.Replace(cname, cnameTemplate, "", -1)

	rcid, err := strconv.ParseUint(s, 10, 64)
	utils.CheckError(err, utils.WarningMode)

	return rcid
}

// Check if a port is already allocated by docker
func isPortAllocatedByDocker(ctx context.Context, port string) bool {

	ctrs := SearchDockerContainers(ctx, "", "", true)
	if ctrs != nil {
		for _, ctr := range ctrs {
			for _, p := range ctr.Ports {
				strp := strconv.FormatUint(uint64(p.PublicPort), 10)

				if strp == port {
					return true
				}
			}
		}
	}

	return false
}

// Get mapped port information of a container
func getContainerPortInfo(ctx context.Context, cname string) (*dockertypes.Port, error) {

	c := SearchDockerContainers(ctx, "name", cname, false)
	if c != nil {
		// TODO: just looking for the first mapped port...
		return &c[0].Ports[0], nil
	}

	return nil, errContainerNotFound
}
