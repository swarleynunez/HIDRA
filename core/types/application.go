package types

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"net"
)

// DCR application model
type Application struct {
	Owner          common.Address
	Info           string   // Encoded application info (ApplicationInfo struct)
	RegisteredAt   *big.Int // Unix time
	UnregisteredAt *big.Int // Unix time
}

type ApplicationInfo struct {
	Description string `json:"desc"`
	IP          net.IP `json:"ip"`   // Virtual service IP
	Protocol    string `json:"prot"` // Virtual service transport protocol (TCP or UDP)
	Port        uint16 `json:"port"` // Virtual service port
}
