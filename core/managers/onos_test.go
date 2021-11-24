package managers

import (
	"context"
	"net"
	"testing"
)

var (
	appid uint64 = 1
	rcid  uint64 = 1
)

func init() {

	InitNode(context.Background(), false)
}

func TestONOSRequest(t *testing.T) {

	ONOSAddVirtualService(appid, "NGINX V1", net.IP("192.168.0.10"), "TCP", 8888)
	ONOSAddVirtualService(appid, "NGINX VX", net.IP("192.168.0.20"), "UDP", 1234)
	ONOSAddVirtualService(appid+1, "NGINX V2", net.IP("192.168.0.11"), "TCP", 8888)
	ONOSAddVirtualService(appid+2, "NGINX V3", net.IP("192.168.0.11"), "TCP", 8080)
	ONOSAddVirtualService(appid+3, "NGINX VX", net.IP("192.168.0.10"), "TCP", 8888)
	ONOSAddVSInstance(context.Background(), appid, rcid, net.IP("172.19.202.107"))
	ONOSAddVSInstance(context.Background(), appid, rcid, net.IP("172.19.202.107"))
	ONOSAddVSInstance(context.Background(), appid, rcid, net.IP("172.19.202.108"))
	ONOSAddVSInstance(context.Background(), appid, rcid, net.IP("172.19.202.109"))
	ONOSDeleteVSInstance(appid, rcid)
	ONOSDeleteVSInstance(appid+1, rcid)
}

/*func TestONOSGetAllVServices(t *testing.T) {

	err := onosc.Request("vss", "")
	utils.CheckError(err, utils.WarningMode)
}

func TestONOSGetAllVServicesOn(t *testing.T) {

	err := onosc.Request("vss_on", "")
	utils.CheckError(err, utils.WarningMode)
}

func TestONOSGetAllVServicesOff(t *testing.T) {

	err := onosc.Request("vss_off", "")
	utils.CheckError(err, utils.WarningMode)
}

func TestONOSActivateVService(t *testing.T) {

	err := onosc.Request("vs_on", "", vsid)
	utils.CheckError(err, utils.WarningMode)
}

func TestONOSGetVService(t *testing.T) {

	err := onosc.Request("vs", "", vsid)
	utils.CheckError(err, utils.WarningMode)
}

func TestONOSDeactivateVService(t *testing.T) {

	err := onosc.Request("vs_off", "", vsid)
	utils.CheckError(err, utils.WarningMode)
}

func TestONOSSetVServiceServer(t *testing.T) {

	s := types.ONOSVSServer{
		IP:       "20.0.0.20",
		Protocol: "UDP",
		Port:     1234,
	}

	err := onosc.Request("server_set", utils.MarshalJSON(s), vsid)
	utils.CheckError(err, utils.WarningMode)
}

func TestONOSAddVServiceInst(t *testing.T) {

	inst := types.ONOSVSInstance{
		ID:       instid + 1,
		IP:       "192.168.0.10",
		Protocol: "UDP",
		Port:     8080,
	}

	err := onosc.Request("inst_add", utils.MarshalJSON(inst), vsid)
	utils.CheckError(err, utils.WarningMode)
}

func TestONOSDeleteVServiceInst(t *testing.T) {

	err := onosc.Request("inst_del", "", vsid, instid)
	utils.CheckError(err, utils.WarningMode)
}

func TestONOSDeleteVService(t *testing.T) {

	err := onosc.Request("vs_del", "", vsid)
	utils.CheckError(err, utils.WarningMode)
}*/
