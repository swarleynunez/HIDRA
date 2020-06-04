package utils

import (
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/swarleynunez/superfog/core/types"
	"os"
	"path"
	"regexp"
)

var (
	errUnknownComp   = errors.New("unknown comparator")
	errMismatchTypes = errors.New("mismatch value types")
	errUnknownType   = errors.New("unknown value type")
)

func FormatPath(paths ...string) (p string) {

	p, err := os.UserHomeDir()
	CheckError(err, WarningMode)

	for _, v := range paths {
		p = path.Join(p, v)
	}

	return
}

func EmptyEthAddress(addr string) bool {

	return addr == new(common.Address).String()
}

func ValidEthAddress(addr string) bool {

	re := regexp.MustCompile("^(?i)(0x)?[0-9a-f]{40}$") // (?i) case insensitive, (0x)? optional hex prefix
	return re.MatchString(addr)
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

func CompareValues(value interface{}, comp types.Comparator, bound interface{}) (r bool, err error) {

	// Uint64 assertion
	if v, ok := value.(uint64); ok {
		if b, ok := bound.(uint64); ok {
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
		if b, ok := bound.(float64); ok {
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
		if b, ok := bound.(string); ok {
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
