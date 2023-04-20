package inputs

import (
	"github.com/swarleynunez/hidra/core/types"
	"net"
)

var AppInfo = types.ApplicationInfo{
	Description: "NGINX Webserver V1.0.0",
	IP:          net.ParseIP("170.100.8.33"),
	Protocol:    "TCP",
	Port:        80,
}
