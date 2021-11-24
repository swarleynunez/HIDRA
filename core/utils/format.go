package utils

import (
	"encoding/json"
	"errors"
	"github.com/docker/distribution/reference"
	"github.com/ethereum/go-ethereum/common"
	"github.com/swarleynunez/superfog/core/types"
	"os"
	"path"
	"regexp"
	"strings"
)

var (
	errUnknownComp   = errors.New("unknown comparator")
	errMismatchTypes = errors.New("value type mismatch")
	errUnknownType   = errors.New("unknown value type")
)

// Docker metadata
var (
	defaultDomain    = "docker.io"
	officialRepoName = "library"
)

////////////////
// Formatters //
////////////////
func FormatPath(paths ...string) (r string) {

	r, err := os.UserHomeDir()
	CheckError(err, FatalMode)

	for i := range paths {
		r = path.Join(r, paths[i])
	}

	return
}

func FormatImageTag(imgTag string) (r string, err error) {

	// Convert tag to full name
	named, err := reference.ParseNormalizedNamed(imgTag)
	if err != nil {
		return
	}

	// Add "latest" tag if no tag found
	named = reference.TagNameOnly(named)

	// Remove domain and repository name from full name
	r = strings.TrimPrefix(named.String(), defaultDomain+"/")
	r = strings.TrimPrefix(r, officialRepoName+"/")

	return
}

//////////////
// Checkers //
//////////////
func EmptyEthAddress(addr string) bool {

	return addr == new(common.Address).String()
}

func ValidEthAddress(addr string) bool {

	re := regexp.MustCompile(`(?i)(0x)?[0-9a-f]{40}`) // (?i) case insensitive, (0x)? optional hex prefix
	return re.MatchString(addr)
}

func CompareValues(value interface{}, comp types.RuleComparator, limit interface{}) (r bool, err error) {

	// Uint64 assertion
	if v, ok := value.(uint64); ok {
		if b, ok := limit.(uint64); ok {
			switch comp {
			case types.EqualComp:
				r = v == b
			case types.NotEqualComp:
				r = v != b
			case types.LessComp:
				r = v < b
			case types.LessOrEqualComp:
				r = v <= b
			case types.GreaterComp:
				r = v > b
			case types.GreaterOrEqualComp:
				r = v >= b
			default:
				err = errUnknownComp
			}
		} else {
			err = errMismatchTypes
		}

		return
	}

	// Float64 assertion
	if v, ok := value.(float64); ok {
		if b, ok := limit.(float64); ok {
			switch comp {
			case types.EqualComp:
				r = v == b
			case types.NotEqualComp:
				r = v != b
			case types.LessComp:
				r = v < b
			case types.LessOrEqualComp:
				r = v <= b
			case types.GreaterComp:
				r = v > b
			case types.GreaterOrEqualComp:
				r = v >= b
			default:
				err = errUnknownComp
			}
		} else {
			err = errMismatchTypes
		}

		return
	}

	// String assertion
	if v, ok := value.(string); ok {
		if b, ok := limit.(string); ok {
			switch comp {
			case types.EqualComp:
				r = v == b
			case types.NotEqualComp:
				r = v != b
			case types.LessComp:
				r = v < b
			case types.LessOrEqualComp:
				r = v <= b
			case types.GreaterComp:
				r = v > b
			case types.GreaterOrEqualComp:
				r = v >= b
			default:
				err = errUnknownComp
			}
		} else {
			err = errMismatchTypes
		}

		return
	}

	err = errUnknownType
	return
}

//////////////
// Encoding //
//////////////
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
