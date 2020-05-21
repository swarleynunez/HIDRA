package utils

import (
	"log"
	"path/filepath"
	"runtime"
)

const (
	InfoMode = iota
	WarningMode
	PanicMode
	FatalMode

	BlueFormat   = "\033[1;34m%s:%d %s\033[0m"
	YellowFormat = "\033[1;33m%s:%d %s\033[0m"
	RedFormat    = "\033[1;31m%s:%d %s\033[0m"
)

func CheckError(err error, mode int) {

	if err != nil {

		// Get file and code line of the error
		_, file, line, _ := runtime.Caller(1)
		file = filepath.Base(file)

		switch mode {

		case InfoMode:
			log.Printf(BlueFormat, file, line, err)
		case WarningMode:
			log.Printf(YellowFormat, file, line, err)
		case PanicMode:
			log.Panicf(RedFormat, file, line, err)
		case FatalMode:
			log.Fatalf(RedFormat, file, line, err)
		}
	}
}
