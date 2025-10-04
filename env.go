package main

import (
	"log"
	"os"
)

func getEnvDefault(key, _default string) (value string) {
	var ok bool
	if value, ok = os.LookupEnv(key); !ok {
		value = _default
	}

	return
}

func getEnvMust(key string) (value string) {
	if value = os.Getenv(key); value == "" {
		log.Fatalf("Environment variable: %s is not provided\n", key)
	}

	return
}
