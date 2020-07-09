package types

// Specification types
type Spec uint8

const (
	CpuSpec Spec = iota
	MemSpec
	DiskSpec
	PktSentSpec
	PktRecvSpec
)

// Rule metric type for each spec
type MetricType uint8

const (
	UnitsMetric MetricType = iota
	PercentMetric
	TagMetric
)

// Rule comparison operators
type Comparator uint8

const (
	EqualComp Comparator = iota
	NotEqualComp
	LessComp
	LessOrEqualComp
	GreaterComp
	GreaterOrEqualComp
)

// Rule actions for the enforcer
type Action uint8

const (
	IgnoreAction    Action = iota
	WarnAction             // Msg to stdout
	LogAction              // Msg to log file
	ProceedAction          // Do it yourself
	SendEventAction        // Ask for help
)

// Rules to check in clients
type Rule struct {
	NameId     string // Unique
	Spec       Spec
	MetricType MetricType
	Comparator Comparator
	Bound      interface{} // uint64, float64 or string
	Action     Action
	Msg        string
}
