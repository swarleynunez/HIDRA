package utils

import "log"

const (
	InfoMode = iota
	WarningMode
	PanicMode
	FatalMode

	BlueFormat   = "\033[1;34m%s\033[0m"
	YellowFormat = "\033[1;33m%s\033[0m"
	RedFormat    = "\033[1;31m%s\033[0m"
)

func CheckError(err error, mode int) {

	if err != nil {

		switch mode {

		case InfoMode:
			log.Printf(BlueFormat, err)
		case WarningMode:
			log.Printf(YellowFormat, err)
		case PanicMode:
			log.Panicf(RedFormat, err)
		case FatalMode:
			log.Fatalf(RedFormat, err)
		}
	}
}
