package onos

import (
	"errors"
	"github.com/swarleynunez/superfog/core/utils"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var (
	errUnauthorized     = errors.New("wrong onos username or password")
	errResourceNotFound = errors.New("onos resource not found")
	errUnsuccessfulReq  = errors.New("unsuccessful onos request")
	errRouteNotFound    = errors.New("onos route not found")
	errParamsMismatch   = errors.New("parameter count mismatch")
)

func (cli *Client) Request(rname, body string, params ...uint64) error {

	// Get route by action name
	if r, found := Routes[rname]; found {
		path, err := parsePath(r.Path, params)
		if err != nil {
			return err
		}

		// Set request using parsed path
		url := cli.BaseURL.String() + path
		req, err := http.NewRequest(r.Method, url, strings.NewReader(body))
		if err != nil {
			return err
		}

		// HTTP headers
		req.SetBasicAuth(utils.GetEnv("ONOS_API_USER"), utils.GetEnv("ONOS_API_PASS"))
		if r.Method == "POST" {
			req.Header.Set("Content-Type", "application/json")
		}
		req.Close = true

		// Send request
		res, err := cli.Client.Do(req)
		if err != nil {
			return err
		}

		// Deferring the response body closure
		defer res.Body.Close()

		// Check HTTP response status code
		switch res.StatusCode {
		case 200:
			// Parse response body
			var rb strings.Builder
			_, err = io.Copy(&rb, res.Body)
			if err != nil {
				return err
			}

			// Handle response body
			r.Handler(rb.String())

			return nil
		case 401:
			return errUnauthorized
		case 404:
			return errResourceNotFound
		default:
			return errUnsuccessfulReq
		}
	}

	return errRouteNotFound
}

func parsePath(path string, params []uint64) (string, error) {

	// Find all parameter template occurrences in path
	re := regexp.MustCompile(`{[^{}]*}`)
	match := re.FindAllString(path, -1)

	// Check parameter count
	if len(params) != len(match) {
		return "", errParamsMismatch
	}

	// Replace parameter template occurrences in path
	for i := range match {
		path = strings.Replace(path, match[i], strconv.FormatUint(params[i], 10), 1)
	}

	return path, nil
}
