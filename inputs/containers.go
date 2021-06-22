package inputs

import (
	"github.com/docker/go-connections/nat"
	"github.com/swarleynunez/superfog/core/types"
)

var Containers = [...]types.ContainerInfo{
	// TODO. Add container commands
	{
		ImageTag: "nginx",
		ContainerSetup: types.ContainerSetup{
			ContainerType: types.ContainerType{
				Impact:      5,
				MainSpec:    types.CpuSpec,
				ServiceType: types.WebServerServ,
			},
			ContainerConfig: types.ContainerConfig{
				CPULimit: 1 * 1e9,
				MemLimit: 512 * 1024 * 1024,
				/*Volumes: []string{
					"nginx-vol:/nginx-vol",
				},*/
				Ports: nat.PortMap{
					"80/tcp": []nat.PortBinding{
						{
							HostPort: "8080",
						},
					},
				},
			},
		},
	},
}
