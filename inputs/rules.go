package inputs

import "github.com/swarleynunez/superfog/core/types"

var Rules = [...]types.Rule{
	{
		NameId:     "rule_1",
		Spec:       types.CpuSpec,
		MetricType: types.PercentMetric,
		Comparator: types.GreaterComp,
		Bound:      float64(33),
		Action:     types.SendEventAction,
		Msg:        "CPU usage % exceeded",
	},
	/*{
		NameId:     "rule_2",
		Spec:       types.MemSpec,
		MetricType: types.PercentMetric,
		Comparator: types.GreaterComp,
		Bound:      float64(75),
		Action:     types.SendEventAction,
		Msg:        "RAM usage % exceeded",
	},
	{
		NameId:     "rule_3",
		Spec:       types.DiskSpec,
		MetricType: types.PercentMetric,
		Comparator: types.GreaterComp,
		Bound:      float64(70),
		Action:     types.SendEventAction,
		Msg:        "Disk space usage % exceeded",
	},*/
}
