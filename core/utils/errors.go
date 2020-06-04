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

	blueFormat   = "\033[1;34m%s:%d %s\033[0m"
	yellowFormat = "\033[1;33m%s:%d %s\033[0m"
	redFormat    = "\033[1;31m%s:%d %s\033[0m"
)

func CheckError(err error, mode int) {

	if err != nil {

		// Get file and code line of the error
		_, file, line, _ := runtime.Caller(1)
		file = filepath.Base(file)

		switch mode {

		case InfoMode:
			log.Printf(blueFormat, file, line, err)
		case WarningMode:
			log.Printf(redFormat, file, line, err)
		case PanicMode: // To recover control
			log.Panicf(redFormat, file, line, err)
		case FatalMode:
			log.Fatalf(redFormat, file, line, err)
		}
	}
}
