package types

import "github.com/ethereum/go-ethereum/common"

type Application struct {
	Owner     common.Address
	Info      string // Encoded
	CreatedAt uint64 // Unix time
	DeletedAt uint64 // Unix time
}

/*type ApplicationInfo struct {
	Ip          string // Virtual service IP
	Protocol    string `json:"proto"` // Virtual service transport protocol (TCP or UDP)
	Port        string // Virtual service port
	Description string `json:"desc"`
}*/
