package inputs

import (
	"github.com/swarleynunez/superfog/core/types"
	"net"
)

var AppInfo = types.ApplicationInfo{
	IP:          net.ParseIP("192.168.0.1"),
	Port:        8080,
	Protocol:    "TCP",
	Description: "NGINX Webapp V1",
}
