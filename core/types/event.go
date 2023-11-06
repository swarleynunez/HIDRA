package types

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/swarleynunez/hidra/core/bindings"
	"math/big"
)

// Event tasks to execute locally
type task uint8

const (
	// About web resources, storage... TODO: IPFS?
	CreateTask task = iota
	ReadTask
	UpdateTask
	DeleteTask

	// About containers
	NewContainerTask
	MigrateContainerTask

	// About nodes
	PingNodeTask
)

// DEL event model
type Event struct {
	EType              string // Encoded event type (EventType struct)
	Sender             common.Address
	Solver             common.Address
	Rcid               uint64 // Optional, depending on the event type
	HasRequiredReplies bool
	HasRequiredVotes   bool
	SentAt             *big.Int // Unix time
	SolvedAt           *big.Int // Unix time
}

type EventType struct {
	RequiredTask task     `json:"task"` // Task to be executed locally by cluster nodes
	Resource     resource `json:"res"`  // Resource used to choose an event solver
}

type EventReply struct {
	Replier   common.Address
	RepScores []bindings.DELReputationScore
	RepliedAt *big.Int // Unix time
}
