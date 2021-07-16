package types

type Rule struct {
	NameId     string // Unique
	Resource   resource
	MetricType RuleMetricType
	Comparator RuleComparator
	Limit      interface{} // uint64, float64 or string
	Action     action
	Msg        string
}

// Rule metric types for each resource
type RuleMetricType uint8

const (
	UnitsMetric RuleMetricType = iota
	PercentMetric
	TagMetric
)

// Rule comparison operators
type RuleComparator uint8

const (
	EqualComp RuleComparator = iota
	NotEqualComp
	LessComp
	LessOrEqualComp
	GreaterComp
	GreaterOrEqualComp
)

// Rule actions for the enforcer
type action uint8

const (
	IgnoreAction    action = iota
	WarnAction             // Msg to stdout
	LogAction              // Msg to log file
	ProceedAction          // Execute a task
	SendEventAction        // Ask for cluster help
)
