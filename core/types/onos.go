package types

import "net/http"

// Client for ONOS virtual service API requests
type ONOSClient struct {
	Scheme   string       // HTTP or HTTPS (currently only HTTP is allowed)
	Host     string       // ONOS controller IP and port to connect to
	BasePath string       // Prepended path to all requests
	Client   *http.Client // To send and receive requests
	Enabled  bool         // Is the ONOS module enabled?
}

// ONOS virtual service model
type ONOSVService struct {
	ID          uint64                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Admin       string                 `json:"admin"`
	State       string                 `json:"state"`
	Server      ONOSVServiceServer     `json:"server"`
	Instances   []ONOSVServiceInstance `json:"instances"`
}

type ONOSVServiceServer struct {
	IP       string `json:"ip"`
	Protocol uint64 `json:"protocol"`
	Port     uint64 `json:"port"`
}

type ONOSVServiceInstance struct {
	ID       uint64 `json:"id"`
	IP       string `json:"ip"`
	Protocol string `json:"protocol"`
	Port     uint64 `json:"port"`
}
