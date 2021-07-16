package inputs

import (
	"github.com/docker/go-connections/nat"
	"github.com/swarleynunez/superfog/core/types"
)

// TODO: add container commands
var CtrInfo = types.ContainerInfo{
	ImageTag: "nginx",
	ContainerType: types.ContainerType{
		ServiceType: types.WebServerServ,
		Impact:      3,
	},
	ContainerConfig: types.ContainerConfig{
		CPULimit: 1 * 1e9,
		MemLimit: 512 * 1024 * 1024,
		Volumes: []string{
			"nginx-vol:/nginx-vol",
		},
		Ports: nat.PortMap{
			"80/tcp": []nat.PortBinding{
				{
					HostPort: "8080",
				},
			},
		},
	},
}
