package types

import (
	"github.com/docker/go-connections/nat"
	"github.com/ethereum/go-ethereum/common"
)

// Container service types
type ServiceType uint8

const (
	ControlServ ServiceType = iota // System control plane management
	OsServ
	WebServerServ // Applications, websites, APIs
	DatabaseServ  // Databases, file storage
	DaemonServ    // Monitoring, proxies, middlewares
	FrameworkServ // Tools, programming languages, compilers
)

// Public and static info (related to the distributed registry)
type Container struct {
	Host       common.Address // Node which runs the container
	Info       string         // Encoded
	StartedAt  uint64         // Unix time
	FinishedAt uint64         // Unix time
}

// General container information
type ContainerInfo struct {
	Id string
	ContainerSetup
	IPAddress string `json:"ip"`
	ImageArch string `json:"arch"`
	ImageOs   string `json:"os"`
	ImageSize uint64 `json:"isize"` // Virtual size (including shared layers)
}

type ContainerSetup struct {
	ContainerType
	ContainerConfig
}

// Dynamic container types
type ContainerType struct {
	Impact      uint8       `json:"impact"` // Importance over the entirely system (0-10)
	MainSpec    Spec        `json:"spec"`   // Mainly used spec
	ServiceType ServiceType `json:"serv"`
}

// Abstraction of all container configs
type ContainerConfig struct {
	ImageTag string `json:"tag"`
	//CPULimit    uint64      `json:"cpu"`   // Maximum CPU quota in nano units to use (0 for unlimited)
	MemLimit    uint64      `json:"mem"`   // Maximum memory to use in bytes (0 for unlimited)
	VolumeBinds []string    `json:"vols"`  // Binding volumes
	Ports       nat.PortMap `json:"ports"` // Binding ports
}

// TODO Desired container state
