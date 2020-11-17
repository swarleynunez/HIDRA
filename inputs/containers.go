package inputs

import (
	"github.com/docker/go-connections/nat"
	"github.com/swarleynunez/superfog/core/types"
)

var Containers = [...]types.ContainerSetup{
	// TODO. Add container commands
	/*{
		ContainerType: types.ContainerType{
			Impact:      1,
			MainSpec:    types.MemSpec,
			ServiceType: types.OsServ,
		},
		ContainerConfig: types.ContainerConfig{
			ImageTag: "alpine",
			//CPULimit: 0.5 * 1e9,
			MemLimit: 1024 * 1024 * 1024,
		},
	},*/
	{
		ContainerType: types.ContainerType{
			Impact:      3,
			MainSpec:    types.CpuSpec,
			ServiceType: types.WebServerServ,
		},
		ContainerConfig: types.ContainerConfig{
			ImageTag: "nginx",
			//CPULimit: 1 * 1e9, TODO
			MemLimit: 512 * 1024 * 1024,
			Ports: nat.PortMap{
				"80/tcp": []nat.PortBinding{
					{
						HostPort: "8080",
					},
				},
			},
		},
	},
	/*{
		ContainerType: types.ContainerType{
			Impact:      5,
			MainSpec:    types.MemSpec,
			ServiceType: types.DatabaseServ,
		},
		ContainerConfig: types.ContainerConfig{
			ImageTag: "mongo",
			//CPULimit: 0.5 * 1e9,
			MemLimit: 1024 * 1024 * 1024,
			Ports: nat.PortMap{
				"27017/tcp": []nat.PortBinding{
					{
						HostPort: "27017",
					},
				},
			},
		},
	},
	{
		ContainerType: types.ContainerType{
			Impact:      7,
			MainSpec:    types.DiskSpec,
			ServiceType: types.DatabaseServ,
		},
		ContainerConfig: types.ContainerConfig{
			ImageTag: "nextcloud",
			//CPULimit: 1 * 1e9,
			MemLimit: 1024 * 1024 * 1024,
			VolumeBinds: []string{
				"nextcloud-vol:/nextcloud-vol",
			},
			Ports: nat.PortMap{
				"80/tcp": []nat.PortBinding{
					{
						HostPort: "8888",
					},
				},
			},
		},
	},*/
}
