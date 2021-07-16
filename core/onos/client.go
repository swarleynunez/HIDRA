package onos

import (
	"errors"
	"github.com/swarleynunez/superfog/core/types"
	"github.com/swarleynunez/superfog/core/utils"
	"net"
	"net/http"
	"os"
	"strconv"
)

var (
	errMalformedIP   = errors.New("malformed IP address")
	errMalformedPort = errors.New("malformed port")
)

func Connect() *types.ONOSClient {

	enabled, err := strconv.ParseBool(os.Getenv("ONOS_ENABLED"))
	utils.CheckError(err, utils.FatalMode)

	if enabled {
		ip := os.Getenv("ONOS_CONTROLLER_IP")
		if net.ParseIP(ip) == nil {
			utils.CheckError(errMalformedIP, utils.FatalMode)
		}

		port, err := strconv.ParseUint(os.Getenv("ONOS_CONTROLLER_PORT"), 10, 16)
		if err != nil || port == 0 {
			utils.CheckError(errMalformedPort, utils.FatalMode)
		}

		strp := strconv.FormatUint(port, 10)
		if !utils.IsPortAvailable("tcp", ip, strp) {
			return &types.ONOSClient{
				Scheme:   "http",
				Host:     net.JoinHostPort(ip, strp),
				BasePath: "/onos/vs",
				Client:   http.DefaultClient,
				Enabled:  true,
			}
		}
	}

	return &types.ONOSClient{}
}
