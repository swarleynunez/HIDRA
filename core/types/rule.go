package types

// Rule metric type for each spec
type MetricType uint8

const (
	UnitsType MetricType = iota
	PercentType
	TagType
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
