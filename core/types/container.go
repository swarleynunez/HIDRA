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
	Host      common.Address // Node which runs the container
	Info      string         // Encoded
	CreatedAt uint64         // Unix time
	DeletedAt uint64         // Unix time
}

// General container information
type ContainerInfo struct {
	//Id              string `json:"id"`
	//ApplicationInfo        // TODO
	ImageTag        string `json:"tag"`
	ImageArch       string `json:"arch"`
	ImageOs         string `json:"os"`
	ImageSize       uint64 `json:"isize"` // Virtual size (including shared layers)
	ContainerSetup
}

// TODO
type ApplicationInfo struct {
	Ip          string // Virtual service IP
	Protocol    string `json:"proto"` // Virtual service transport protocol (TCP or UDP)
	Port        string // Virtual service port
	Description string `json:"desc"`
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
	CPULimit uint64      `json:"cpu"`   // Maximum CPU quota in nano units to use (0 for unlimited)
	MemLimit uint64      `json:"mem"`   // Maximum memory to use in bytes (0 for unlimited)
	Volumes  []string    `json:"vols"`  // Binding volumes
	Ports    nat.PortMap `json:"ports"` // Binding ports
}
