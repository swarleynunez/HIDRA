package types

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"net"
)

// Node resource types
type resource uint8

const (
	NoResource resource = iota
	CpuResource
	MemResource
	DiskResource
	PktSentResource
	PktRecvResource
	AllResources
)

// DDR node model
type NodeData struct {
	Controller   common.Address
	Specs        string // Encoded (NodeSpecs struct)
	Reputation   int64
	RegisteredAt *big.Int // Unix time
}

// Node "static" specifications
type NodeSpecs struct {
	Arch      string  `json:"arch"`
	Cores     uint64  `json:"cores,"`     // Logical cores number
	CpuMhz    float64 `json:"mhz,string"` // Physical cores frequency
	MemTotal  uint64  `json:"mem"`        // In bytes
	DiskTotal uint64  `json:"disk"`       // In bytes
	OS        string  `json:"os"`
	IP        net.IP  `json:"ip"`
	//Location  NodeLocation `json:"loc"`
}

/*type NodeLocation struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"long"`
}*/
