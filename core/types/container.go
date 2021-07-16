package types

import (
	"github.com/docker/go-connections/nat"
	"math/big"
)

// Container service types
type serviceType uint8

const (
	ControlServ serviceType = iota // System control plane management
	OsServ
	WebServerServ // Applications, websites, APIs
	DatabaseServ  // Databases, file storage
	DaemonServ    // Monitoring, proxies, middlewares
	FrameworkServ // Tools, programming languages, compilers
)

// DCR container model
type Container struct {
	Appid          uint64
	Info           string // Encoded container info (ContainerInfo struct)
	Autodeployed   bool
	RegisteredAt   *big.Int // Unix time
	UnregisteredAt *big.Int // Unix time
}

type ContainerInfo struct {
	ImageTag string `json:"itag"`
	ContainerType
	ContainerConfig
}

type ContainerType struct {
	ServiceType serviceType `json:"type"`
	Impact      uint8       `json:"impact"` // Importance over the entirely system (0-10)
}

// Abstraction of all container configs
type ContainerConfig struct {
	CPULimit uint64      `json:"lcpu"`    // Maximum CPU quota in nano units to use (0 for unlimited)
	MemLimit uint64      `json:"lmem"`    // Maximum memory to use in bytes (0 for unlimited)
	Volumes  []string    `json:"volumes"` // Binding volumes
	Ports    nat.PortMap `json:"ports"`   // Binding ports
}
