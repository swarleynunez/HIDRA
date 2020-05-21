package utils

import (
	"encoding/json"
	"os"
	"path"
	"regexp"
)

func FormatPath(paths ...string) (p string) {

	p, err := os.UserHomeDir()
	CheckError(err, WarningMode)

	for _, v := range paths {
		p = path.Join(p, v)
	}

	return
}

func ValidEthAddress(addr string) (r bool) {

	re := regexp.MustCompile("^(?i)(0x)?[0-9a-f]{40}$") // (?i) case insensitive, (0x)? optional hex prefix
	r = re.MatchString(addr)

	return
}

func MarshalJSON(v interface{}) string {

	// Encode any struct to JSON
	bytes, err := json.Marshal(v)
	CheckError(err, WarningMode)

	return string(bytes)
}

func UnmarshalJSON(data string, v interface{}) {

	// String to bytes slice
	bytes := []byte(data)

	// Decode JSON to any struct
	err := json.Unmarshal(bytes, v)
	CheckError(err, WarningMode)
}
