package inputs

import (
	"github.com/docker/go-connections/nat"
	"github.com/swarleynunez/hidra/core/types"
)

// TODO: add Docker container commands
var CtrInfo = types.ContainerInfo{
	ImageTag: "nginx",
	ContainerType: types.ContainerType{
		ServiceType: types.WebServerServ,
		Impact:      3,
	},
	ContainerConfig: types.ContainerConfig{
		CpuLimit: 1 * 1e9,
		MemLimit: 1024 * 1024 * 1024,
		Envs:     []string{"env1=hi", "env2=bye"},
		Volumes:  []string{"nginx-vol:/nginx-vol"},
		Ports: nat.PortMap{
			"80/tcp": []nat.PortBinding{
				{
					HostPort: "8888",
				},
			},
		},
	},
}
