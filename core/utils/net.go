package utils

import "net"

func IsPortAvailable(network, host, port string) bool {

	conn, err := net.Dial(network, net.JoinHostPort(host, port))

	if err == nil && conn != nil {
		// Successful connection
		conn.Close()
		return false
	} else {
		return true
	}
}
