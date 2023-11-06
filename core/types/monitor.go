package types

// MonitorV1
type CycleCounter struct { // Rule cycle counter (rcc)
	Measures uint64
	Triggers uint64
}

// MonitorV2
type PacketCounter struct {
	Sent  uint64
	Recv  uint64
	Total uint64
	Max   uint64
}

// Experiments
type EventTimes struct {
	Start int64
	End   int64
}
