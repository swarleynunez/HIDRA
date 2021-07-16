package managers

import (
	"context"
	"github.com/swarleynunez/superfog/core/utils"
	"testing"
)

func init() {

	InitNode(context.Background(), false)
}

/*func TestConn(t *testing.T) {

	baseURL := url.URL{
		Scheme: "http",
		Host:   "192.168.0.41:8181",
		Path:   "/onos/vs",
	}

	type Route struct {
		Method string
		Path   string
		//Handler func()
	}

	var routes = map[string]Route{}
	routes["all"] = Route{
		Method: "GET",
		Path:   "/all",
	}

	req, err := http.NewRequest(routes["all"].Method, baseURL.String()+routes["all"].Path, nil)
	utils.CheckError(err, utils.WarningMode)

	req.SetBasicAuth("onos", "rocks")

	res, err := http.DefaultClient.Do(req)
	utils.CheckError(err, utils.WarningMode)

	// Build a string from the response body
	var body strings.Builder
	_, err = io.Copy(&body, res.Body)
	utils.CheckError(err, utils.WarningMode)

	var vss struct{ VServices []types.ONOSVService }
	utils.UnmarshalJSON(body.String(), &vss)

	t.Log(vss)
}*/

func TestONOSConnect(t *testing.T) {

	//t.Log(onos.Connect())
	t.Log(utils.GetEnv("ONOS_ENABLED"))
}
