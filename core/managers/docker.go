package managers

import (
	"context"
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

var (
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

	out, err := _dcli.ImagePull(ctx, imgTag, dockertypes.ImagePullOptions{})
	utils.CheckError(err, utils.WarningMode)
	_, err = io.Copy(ioutil.Discard, out) // Discard output to /dev/null
	utils.CheckError(err, utils.WarningMode)
}

////////////////
// Containers //
////////////////
func createContainer(ctx context.Context, cc *types.ContainerConfig) string {

	if !existImageLocally(ctx, cc.ImageTag) {
		pullImage(ctx, cc.ImageTag)
	}

	// Set configs
	ctrConfig := &container.Config{Image: cc.ImageTag}
	hostConfig := &container.HostConfig{
		Binds:        cc.VolumeBinds,
		PortBindings: cc.Ports,
		Resources: container.Resources{
			Memory:   int64(cc.MemLimit),
			NanoCPUs: int64(cc.CPULimit),
		},
	}
	netConfig := &network.NetworkingConfig{}

	resp, err := _dcli.ContainerCreate(ctx, ctrConfig, hostConfig, netConfig, nil, "")
	utils.CheckError(err, utils.WarningMode)

	return resp.ID
}

func startContainer(ctx context.Context, cid string, ctype *types.ContainerType) {

	err := _dcli.ContainerStart(ctx, cid, dockertypes.ContainerStartOptions{})
	utils.CheckError(err, utils.WarningMode)

	// Record container on distributed registry
	cinfo := getContainerInfo(ctx, cid)
	cinfo.ContainerType = *ctype
	stime := getContainerStartTime(ctx, cid)
	recordContainerOnReg(cinfo, stime, cid)
}

func stopContainer(ctx context.Context, cname string) {

	// SIGTERM instead of SIGKILL
	err := _dcli.ContainerStop(ctx, cname, nil) // nil = engine default timeout
	utils.CheckError(err, utils.WarningMode)
}

func removeContainer(ctx context.Context, cname string) {

	// Remove container from distributed registry
	rcid := getRegContainerId(cname)
	ftime := getContainerFinishTime(ctx, cname)
	removeContainerFromReg(rcid, ftime)

	err := _dcli.ContainerRemove(ctx, cname, dockertypes.ContainerRemoveOptions{})
	utils.CheckError(err, utils.WarningMode)

	// Remove unused volumes
	pruneVolumes(ctx)
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
func SetContainerName(ctx context.Context, cid string, rcid uint64) (cname string) {

	// Format container name
	cname = cnameTemplate + strconv.FormatUint(rcid, 10)

	err := _dcli.ContainerRename(ctx, cid, cname)
	utils.CheckError(err, utils.WarningMode)

	return
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

func getContainerFinishTime(ctx context.Context, cname string) uint64 {

	// Get container
	ctr, err := _dcli.ContainerInspect(ctx, cname)
	utils.CheckError(err, utils.WarningMode)

	// Get finish unix time
	ftime, err := time.Parse(time.RFC3339, ctr.State.FinishedAt)
	utils.CheckError(err, utils.WarningMode)

	return uint64(ftime.Unix())
}
