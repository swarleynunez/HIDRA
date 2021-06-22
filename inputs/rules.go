package inputs

import "github.com/swarleynunez/superfog/core/types"

var Rules = [...]types.Rule{
	{
		NameId:     "rule_1",
		Spec:       types.CpuSpec,
		MetricType: types.PercentMetric,
		Comparator: types.GreaterComp,
		Bound:      float64(1),
		Action:     types.SendEventAction,
		Msg:        "CPU usage % exceeded",
	},
	{
		NameId:     "rule_2",
		Spec:       types.MemSpec,
		MetricType: types.PercentMetric,
		Comparator: types.GreaterComp,
		Bound:      float64(1),
		Action:     types.IgnoreAction,
		Msg:        "RAM usage % exceeded",
	},
	{
		NameId:     "rule_3",
		Spec:       types.DiskSpec,
		MetricType: types.PercentMetric,
		Comparator: types.GreaterComp,
		Bound:      float64(1),
		Action:     types.IgnoreAction,
		Msg:        "Disk space usage % exceeded",
	},
	{
		NameId:     "rule_4",
		Spec:       types.PktSentSpec,
		MetricType: types.UnitsMetric,
		Comparator: types.GreaterComp,
		Bound:      uint64(1),
		Action:     types.IgnoreAction,
		Msg:        "Sent packet limit exceeded",
	},
	{
		NameId:     "rule_5",
		Spec:       types.PktRecvSpec,
		MetricType: types.UnitsMetric,
		Comparator: types.GreaterComp,
		Bound:      uint64(1),
		Action:     types.IgnoreAction,
		Msg:        "Received packet limit exceeded",
	},
}
