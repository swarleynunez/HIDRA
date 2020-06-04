package policies

import "github.com/swarleynunez/superfog/core/types"

var Rules = [...]types.Rule{
	{
		NameId:     "rule_1",
		Spec:       types.CpuSpec,
		MetricType: types.PercentType,
		Comparator: types.GreaterComp,
		Bound:      float64(50),
		Action:     types.WarnAction,
		Msg:        "CPU usage % exceeded",
	},
	{
		NameId:     "rule_2",
		Spec:       types.MemSpec,
		MetricType: types.PercentType,
		Comparator: types.GreaterComp,
		Bound:      float64(55),
		Action:     types.WarnAction,
		Msg:        "RAM usage % exceeded",
	},
	{
		NameId:     "rule_3",
		Spec:       types.DiskSpec,
		MetricType: types.PercentType,
		Comparator: types.GreaterComp,
		Bound:      float64(60),
		Action:     types.WarnAction,
		Msg:        "Disk space usage % exceeded",
	},
}
