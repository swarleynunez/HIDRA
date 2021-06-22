package managers

import (
	"context"
	"fmt"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/swarleynunez/superfog/core/types"
	"github.com/swarleynunez/superfog/core/utils"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

const (
	cnameTemplate = "registry_ctr_"
)

////////////
// Images //
////////////
func existImageLocally(ctx context.Context, imgTag string) bool {

	// Check and format tag
	imgTag, err := utils.FormatImageTag(imgTag)
	utils.CheckError(err, utils.WarningMode)

	// Get all local images
	images, err := _dcli.ImageList(ctx, dockertypes.ImageListOptions{All: true})
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

	fmt.Printf(blueMsgFormat, time.Now().Format("15:04:05.000000"), "Downloading '"+imgTag+"' image...")
	out, err := _dcli.ImagePull(ctx, imgTag, dockertypes.ImagePullOptions{})
	utils.CheckError(err, utils.WarningMode)
	_, err = io.Copy(ioutil.Discard, out) // Discard output to /dev/null
	utils.CheckError(err, utils.WarningMode)
}

////////////////
// Containers //
////////////////
func createDockerContainer(ctx context.Context, cinfo *types.ContainerInfo, cname string) string {

	// Check and format image tag
	imgTag, err := utils.FormatImageTag(cinfo.ImageTag)
	utils.CheckError(err, utils.WarningMode)

	if !existImageLocally(ctx, imgTag) {
		pullImage(ctx, imgTag)
	}

	ports := checkNodePorts(ctx, cinfo.Ports)

	// Set configs
	ctrConfig := &container.Config{Image: imgTag}
	hostConfig := &container.HostConfig{
		Binds:        cinfo.Volumes,
		PortBindings: ports,
		Resources: container.Resources{
			Memory:   int64(cinfo.MemLimit),
			NanoCPUs: int64(cinfo.CPULimit),
		},
	}
	netConfig := &network.NetworkingConfig{}

	resp, err := _dcli.ContainerCreate(ctx, ctrConfig, hostConfig, netConfig, nil, cname)
	utils.CheckError(err, utils.WarningMode)

	return resp.ID
}

func startDockerContainer(ctx context.Context, ctrNameId string) { // ctrNameId = id or name

	err := _dcli.ContainerStart(ctx, ctrNameId, dockertypes.ContainerStartOptions{})
	utils.CheckError(err, utils.WarningMode)
}

func restartDockerContainer(ctx context.Context, cname string) {

	// Stop and start container
	err := _dcli.ContainerRestart(ctx, cname, nil) // nil = do not wait to start container
	utils.CheckError(err, utils.WarningMode)
}

func stopDockerContainer(ctx context.Context, cname string) {

	// SIGTERM instead of SIGKILL
	err := _dcli.ContainerStop(ctx, cname, nil) // nil = engine default timeout
	utils.CheckError(err, utils.WarningMode)
}

func removeDockerContainer(ctx context.Context, cname string) {

	// TODO. Forcing the container remove
	err := _dcli.ContainerRemove(ctx, cname, dockertypes.ContainerRemoveOptions{Force: true})
	utils.CheckError(err, utils.WarningMode)

	// TODO. Remove unused volumes
	//pruneVolumes(ctx)
}

// all: only running containers (false) or all containers (true)
func SearchDockerContainers(ctx context.Context, key, value string, all bool) *[]dockertypes.Container {

	filter := filters.Args{}
	if key != "" && value != "" {
		filter = filters.NewArgs(filters.KeyValuePair{Key: key, Value: value})
	}

	ctr, err := _dcli.ContainerList(ctx, dockertypes.ContainerListOptions{Size: true, All: all, Filters: filter})
	utils.CheckError(err, utils.WarningMode)

	if len(ctr) > 0 {
		return &ctr
	} else {
		return nil
	}
}

/////////////
// Volumes //
/////////////
func pruneVolumes(ctx context.Context) {

	_, err := _dcli.VolumesPrune(ctx, filters.Args{})
	utils.CheckError(err, utils.WarningMode)
}

/////////////
// Helpers //
/////////////
func SetContainerName(ctx context.Context, cid, cname string) {

	err := _dcli.ContainerRename(ctx, cid, cname)
	utils.CheckError(err, utils.WarningMode)
}

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

func getContainerStartTime(ctx context.Context, cid string) uint64 {

	// Get container
	ctr, err := _dcli.ContainerInspect(ctx, cid)
	utils.CheckError(err, utils.WarningMode)

	// Get start unix time
	stime, err := time.Parse(time.RFC3339, ctr.State.StartedAt)
	utils.CheckError(err, utils.WarningMode)

	return uint64(stime.Unix())
}

// Check if a port is already allocated by docker
func isPortAllocatedByDocker(ctx context.Context, port string) bool {

	ctrs := SearchDockerContainers(ctx, "", "", true)
	if ctrs != nil {
		for _, ctr := range *ctrs {
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
