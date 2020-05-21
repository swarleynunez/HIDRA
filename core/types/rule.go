package types

// Rules to check in clients
type Rule struct {
	Spec                   string // cpu, mem, disk...
	Min, Max               uint64
	MinPercent, MaxPercent string
	InfoMsg                string
	Action                 string // ignore, warning, proceed, send event
}
