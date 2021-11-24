package onos

import (
	"errors"
	"github.com/swarleynunez/superfog/core/utils"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var (
	errMalformedIP   = errors.New("malformed onos controller ip")
	errMalformedPort = errors.New("malformed onos controller port")
)

// Client for ONOS virtual service API requests
type Client struct {
	BaseURL *url.URL     // Base URL to all requests
	Client  *http.Client // To send and receive requests
	Enabled bool         // Is the ONOS module enabled?
}

func Connect() *Client {

	enabled, err := strconv.ParseBool(utils.GetEnv("ONOS_ENABLED"))
	if err != nil {
		enabled = false
	}

	if enabled {
		ip := utils.GetEnv("ONOS_CONTROLLER_IP")
		if net.ParseIP(ip) == nil {
			utils.CheckError(errMalformedIP, utils.FatalMode)
		}

		port, err := strconv.ParseUint(utils.GetEnv("ONOS_CONTROLLER_PORT"), 10, 16)
		if err != nil || port == 0 {
			utils.CheckError(errMalformedPort, utils.FatalMode)
		}

		// Set ONOS client
		onosc := &Client{
			BaseURL: &url.URL{
				Scheme: "http",
				Host:   net.JoinHostPort(ip, strconv.FormatUint(port, 10)),
				Path:   "/onos/vs",
			},
			Client: &http.Client{
				Timeout: 30 * time.Second,
			},
			Enabled: true,
		}

		// Initialize ONOS API routes
		initRoutes()

		// Check connection
		err = onosc.Request("ping", "")
		utils.CheckError(err, utils.FatalMode)

		return onosc
	}

	return &Client{}
}
