package types

import "github.com/ethereum/go-ethereum/common"

// Event tasks for run
type Task uint8

const (
	// About web resources, storage... TODO: IPFS?
	CreateTask Task = iota
	ReadTask
	UpdateTask
	DeleteTask

	// About containers
	NewContainerTask
	MigrateContainerTask

	// About nodes
	PingNodeTask
	RequestResourcesTask // TODO: to update rules dynamically
)

// Network nodes events
type Event struct {
	DynType  string // Encoded dynamic event type
	Sender   common.Address
	Solver   common.Address
	SentAt   uint64 // Unix time
	SolvedAt uint64 // Unix time
}

// Dynamic event types
type EventType struct {	// TODO: delete
	Spec Spec `json:"spec"` // Problematic spec
	Task Task `json:"task"`
	// TODO
	Metadata map[string]interface{} `json:"meta"` // Realtime metadata (container id, suggestions to solver)
}

// Network nodes replies to an event
type Reply struct {
	Replier   common.Address
	NodeState string // Encoded node state
	Voters    []common.Address
	RepliedAt uint64 // Unix time
}
