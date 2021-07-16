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
	IP          net.IP `json:"ip"`   // Virtual service IP
	Port        uint16 `json:"port"` // Virtual service port
	Protocol    string `json:"prot"` // Virtual service transport protocol (TCP or UDP)
	Description string `json:"desc"`
}
