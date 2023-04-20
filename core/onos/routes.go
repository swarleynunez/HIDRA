package onos

import (
	"errors"
	"github.com/swarleynunez/hidra/core/utils"
)

var (
	errDuplicatedRoute = errors.New("duplicated onos route")
)

// ONOS virtual service API routes
type Route struct {
	Method  string            // HTTP method
	Path    string            // Endpoint path
	Handler func(body string) // Response handler
}

// Routes mapped by route name
var Routes = map[string]Route{}

func initRoutes() {

	// Named routes. Parameters (any name) between "{" and "}"
	get("ping", "/", func(body string) {})
	get("vss", "/all", func(body string) {
		//var vss struct{ VServices []types.ONOSVirtualService }
		//utils.UnmarshalJSON(body, &vss)
		//fmt.Println(vss)
	})
	get("vss_on", "/on", func(body string) {
		//var vss struct{ VServicesON []types.ONOSVirtualService }
		//utils.UnmarshalJSON(body, &vss)
		//fmt.Println(vss)
	})
	get("vss_off", "/off", func(body string) {
		//var vss struct{ VServicesOFF []types.ONOSVirtualService }
		//utils.UnmarshalJSON(body, &vss)
		//fmt.Println(vss)
	})
	get("vs", "/{vs_id}", func(body string) {
		//var vs struct{ VService types.ONOSVirtualService }
		//utils.UnmarshalJSON(body, &vs)
		//fmt.Println(vs)
	})
	post("vs_add", "/add", func(body string) {})
	get("vs_on", "/{vs_id}/on", func(body string) {})
	get("vs_off", "/{vs_id}/off", func(body string) {})
	//post("server_set", "/{vs_id}/setserver", func(body string) {})
	post("inst_add", "/{vs_id}/addinstance", func(body string) {})
	get("inst_del", "/{vs_id}/{inst_id}/del", func(body string) {})
	get("vs_del", "/{vs_id}/del", func(body string) {})
}

func get(rname, path string, handler func(body string)) {

	if _, found := Routes[rname]; !found {
		Routes[rname] = Route{Method: "GET", Path: path, Handler: handler}
	} else {
		utils.CheckError(errDuplicatedRoute, utils.FatalMode)
	}
}

func post(rname, path string, handler func(body string)) {

	if _, found := Routes[rname]; !found {
		Routes[rname] = Route{Method: "POST", Path: path, Handler: handler}
	} else {
		utils.CheckError(errDuplicatedRoute, utils.FatalMode)
	}
}
