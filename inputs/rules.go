package inputs

import "github.com/swarleynunez/superfog/core/types"

var Rules = [...]types.Rule{
	{
		NameId:     "rule_1",
		Resource:   types.CpuResource,
		MetricType: types.PercentMetric,
		Comparator: types.GreaterComp,
		Limit:      float64(1),
		Action:     types.SendEventAction,
		Msg:        "CPU usage % exceeded",
	},
	{
		NameId:     "rule_2",
		Resource:   types.MemResource,
		MetricType: types.PercentMetric,
		Comparator: types.GreaterComp,
		Limit:      float64(1),
		Action:     types.IgnoreAction,
		Msg:        "RAM usage % exceeded",
	},
	{
		NameId:     "rule_3",
		Resource:   types.DiskResource,
		MetricType: types.PercentMetric,
		Comparator: types.GreaterComp,
		Limit:      float64(1),
		Action:     types.IgnoreAction,
		Msg:        "Disk space usage % exceeded",
	},
	{
		NameId:     "rule_4",
		Resource:   types.PktSentResource,
		MetricType: types.UnitsMetric,
		Comparator: types.GreaterComp,
		Limit:      uint64(1),
		Action:     types.IgnoreAction,
		Msg:        "Sent packet limit exceeded",
	},
	{
		NameId:     "rule_5",
		Resource:   types.PktRecvResource,
		MetricType: types.UnitsMetric,
		Comparator: types.GreaterComp,
		Limit:      uint64(1),
		Action:     types.IgnoreAction,
		Msg:        "Received packet limit exceeded",
	},
}
