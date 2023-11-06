package types

import (
	"github.com/ethereum/go-ethereum/common"
)

type NodeStore map[common.Address]*NodeInfo

type NodeInfo struct {
	CurrentEpoch EpochInfo
	Reputation   ReputationInfo
}

type EpochInfo struct {
	OKPackets    uint64
	TotalPackets uint64
	Latencies    []uint64
}

type ReputationInfo struct {
	Values []uint8
	Score  float64
}

type ReputationScoreCounter struct {
	Scores []float64
	Total  float64
}
