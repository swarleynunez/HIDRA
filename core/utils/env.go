package utils

import (
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"os"
)

func LoadEnv() {

	// Load .env keys for this process
	err := godotenv.Load("../../.env") // TODO: testing
	CheckError(err, FatalMode)
}

func GetEnv(key string) (value string) {

	value, found := os.LookupEnv(key)
	if !found || value == "" {
		CheckError(errors.New(fmt.Sprintf("\"%s\" environment variable not set", key)), FatalMode)
	}

	return
}

func SetEnvKey(key, value string) {

	// Read .env keys into a map
	env, err := godotenv.Read("../../.env") // TODO: testing
	CheckError(err, WarningMode)

	// Add or modify a key-value
	env[key] = value

	// Write map into .env file
	err = godotenv.Write(env, "../../.env") // TODO: testing
	CheckError(err, WarningMode)

	// Reload .env configuration
	overloadEnv()
}

func overloadEnv() {

	// Reload .env keys for this process
	err := godotenv.Overload("../../.env") // TODO: testing
	CheckError(err, WarningMode)
}
