package types

// ONOS virtual service model
type ONOSVirtualService struct {
	ID          uint64           `json:"id"`
	Description string           `json:"description"`
	State       string           `json:"state,omitempty"`
	Server      ONOSVSServer     `json:"server"`
	Instances   []ONOSVSInstance `json:"instances"`
}

type ONOSVSServer struct {
	IP       string `json:"ip"`
	Protocol string `json:"protocol"`
	Port     uint16 `json:"port"`
}

type ONOSVSInstance struct {
	ID       uint64 `json:"id"`
	IP       string `json:"ip"`
	Protocol string `json:"protocol"`
	Port     uint16 `json:"port"`
}
