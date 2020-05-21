package types

import "github.com/ethereum/go-ethereum/common"

// Network nodes events
type Event struct {
	DynType   string // Encoded dynamic event type
	Sender    common.Address
	CreatedAt uint64 // Unix time
	Solver    common.Address
	SolvedAt  uint64 // Unix time
}

// Network nodes replies to an event
type Reply struct {
	Sender    common.Address
	NodeState string // Encoded node state
	CreatedAt uint64 // Unix time
	Voters    []common.Address
}

// Dynamic event types. To send within events
type EventType struct {
	Spec     string            `json:"spec"` // cpu, mem, disk...
	Task     string            `json:"task"` // create, read, update, delete, migrate, kill, ping...
	Metadata map[string]string `json:"meta"` // Present and future metadata
}
