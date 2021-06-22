package utils

import (
	"github.com/joho/godotenv"
)

func LoadEnv() {

	// Load .env keys for this process
	err := godotenv.Load()
	CheckError(err, FatalMode)
}

func SetEnvKey(key, value string) {

	// Read .env keys into a map
	env, err := godotenv.Read()
	CheckError(err, WarningMode)

	// Add or modify a key-value
	env[key] = value

	// Write map into .env file
	err = godotenv.Write(env, ".env")
	CheckError(err, WarningMode)

	// Reload .env configuration
	overloadEnv()
}

func overloadEnv() {

	// Reload .env keys for this process
	err := godotenv.Overload()
	CheckError(err, WarningMode)
}
