package utils

import (
	"os"
	"path"
	"regexp"
	"strconv"
)

const (
	HttpMode = iota
	HttpsMode
)

func FormatUrl(ip string, port, mode int) (url string) {

	switch mode {

	case HttpMode:
		url = "http://" + ip + ":" + strconv.Itoa(port)
	case HttpsMode:
		url = "https://" + ip + ":" + strconv.Itoa(port)
	}

	return
}

func FormatPath(paths ...string) (p string) {

	p, err := os.UserHomeDir()
	CheckError(err, FatalMode)

	for _, v := range paths {
		p = path.Join(p, v)
	}

	return
}

func CheckEthAddress(address string) bool {

	re := regexp.MustCompile("^(?i)(0x)?[0-9a-f]{40}$") // (?i) case insensitive, (0x)? optional hex prefix

	return re.MatchString(address)
}
