package types

// Node static info
type NodeSpecs struct {
	Arch       string  `json:"arch"`
	Cores      uint64  `json:"cores"`      // Logical cores number
	CpuMhz     float64 `json:"mhz,string"` // Physical cores frequency
	MemTotal   uint64  `json:"mem"`        // In bytes
	DiskTotal  uint64  `json:"disk"`       // In bytes
	FileSystem string  `json:"fs"`
	Os         string  `json:"os"`
	Hostname   string  `json:"hostname"`
	BootTime   uint64  `json:"boot"` // Unix time
	// TODO: add IP?
	// TODO: add latitude and longitude?
	// TODO: delete hostname?
}
