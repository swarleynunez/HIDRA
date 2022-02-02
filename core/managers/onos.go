package managers

import (
	"context"
	"github.com/swarleynunez/superfog/core/types"
	"github.com/swarleynunez/superfog/core/utils"
	"net"
	"strings"
)

func ONOSAddVirtualService(appid uint64, desc string, vip net.IP, vproto string, vport uint16) {

	if !_onosc.Enabled {
		return
	}

	vs := types.ONOSVirtualService{
		ID:          appid,
		Description: desc,
		Server: types.ONOSVSServer{ // VS server (virtual fields)
			IP:       vip.String(),
			Protocol: strings.ToUpper(vproto),
			Port:     vport,
		},
	}

	err := _onosc.Request("vs_add", utils.MarshalJSON(vs))
	utils.CheckError(err, utils.WarningMode)
}

func ONOSActivateVirtualService(appid uint64) {

	if !_onosc.Enabled {
		return
	}

	err := _onosc.Request("vs_on", "", appid)
	utils.CheckError(err, utils.WarningMode)
}

/*func ONOSDeactivateVirtualService(appid uint64) {

	if !_onosc.Enabled {
		return
	}

	err := _onosc.Request("vs_off", "", appid)
	utils.CheckError(err, utils.WarningMode)
}*/

func ONOSDeleteVirtualService(appid uint64) {

	if !_onosc.Enabled {
		return
	}

	err := _onosc.Request("vs_del", "", appid)
	utils.CheckError(err, utils.WarningMode)
}

func ONOSAddVSInstance(ctx context.Context, appid, rcid uint64, nip net.IP) {

	if !_onosc.Enabled {
		return
	}

	port, err := getContainerPortInfo(ctx, GetContainerName(rcid))
	utils.CheckError(err, utils.WarningMode)

	if err == nil {
		inst := types.ONOSVSInstance{
			ID:       rcid,
			IP:       nip.String(),
			Protocol: strings.ToUpper(port.Type),
			Port:     port.PublicPort,
		}

		err = _onosc.Request("inst_add", utils.MarshalJSON(inst), appid)
		utils.CheckError(err, utils.WarningMode)
	}
}

func ONOSDeleteVSInstance(appid, rcid uint64) {

	if !_onosc.Enabled {
		return
	}

	err := _onosc.Request("inst_del", "", appid, rcid)
	utils.CheckError(err, utils.WarningMode)
}
