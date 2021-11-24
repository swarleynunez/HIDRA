package inputs

import (
	"github.com/swarleynunez/superfog/core/types"
	"net"
)

var AppInfo = types.ApplicationInfo{
	Description: "POSTGRESQL Database V1.0.0",
	IP:          net.ParseIP("170.100.8.33"),
	Protocol:    "TCP",
	Port:        80,
}
