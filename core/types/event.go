package types

import "github.com/ethereum/go-ethereum/common"

// Event tasks for run
type Task uint8

const (
	CreateTask Task = iota
	ReadTask
	UpdateTask
	DeleteTask
	ReloadTask
	StartTask
	StopTask
	RestartTask
	MigrateTask
	KillTask
	PingTask
	GetTask  // Downloading
	PostTask // Uploading
)

// Network nodes events
type Event struct {
	DynType   string // Encoded dynamic event type
	Sender    common.Address
	CreatedAt uint64 // Unix time
	Solver    common.Address
	SolvedAt  uint64 // Unix time
}

// Dynamic event types
type EventType struct {
	Spec     Spec                   `json:"spec"` // Problematic spec
	Task     Task                   `json:"task"`
	Metadata map[string]interface{} `json:"meta"` // Realtime metadata (container id, suggestions to solver)
}

// Network nodes replies to an event
type Reply struct {
	Sender    common.Address
	NodeState string // Encoded node state
	CreatedAt uint64 // Unix time
	Voters    []common.Address
}
