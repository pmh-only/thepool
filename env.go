package main

import (
	"log"
	"os"
	"strconv"
)

const MB = int64(1024 * 1024)

var WEBSERVER_PORT = getEnvDefault("WEBSERVER_PORT", "8080")
var WEBSERVER_TLS_CERT_PATH = getEnvDefault("WEBSERVER_TLS_CERT_PATH", "/tmp/cert.pem")
var WEBSERVER_TLS_KEY_PATH = getEnvDefault("WEBSERVER_TLS_KEY_PATH", "/tmp/key.pem")
var WEBSERVER_SIZE_LIMIT_MB = getEnvDefaultInt("WEBSERVER_SIZE_LIMIT_MB", "128")
var WEBSERVER_STATIC_ENDPOINT_PREFIX = getEnvMust("WEBSERVER_STATIC_ENDPOINT_PREFIX")

var MINIO_ENDPOINT = getEnvMust("MINIO_ENDPOINT")
var MINIO_ACCESS_KEY_ID = getEnvMust("MINIO_ACCESS_KEY_ID")
var MINIO_SECRET_ACCESS_KEY = getEnvMust("MINIO_SECRET_ACCESS_KEY")
var MINIO_BUCKET_NAME = getEnvMust("MINIO_BUCKET_NAME")
var MINIO_BUCKET_KEY_PREFIX = getEnvDefault("MINIO_BUCKET_KEY_PREFIX", "")
var MINIO_ENDPOINT_SECURE = getEnvDefault("MINIO_ENDPOINT_SECURE", "true") != "false"

var CRONJOB_POOL_SIZE_LIMIT_MB = getEnvDefaultInt("CRONJOB_POOL_SIZE_LIMIT_MB", "1024")

var DATABASE_USER = getEnvMust("DATABASE_USER")
var DATABASE_PASSWORD = getEnvDefault("DATABASE_PASSWORD", "")
var DATABASE_SCHEMA = getEnvMust("DATABASE_SCHEMA")
var DATABASE_HOST = getEnvMust("DATABASE_HOST")
var DATABASE_PORT = getEnvMust("DATABASE_PORT")

var DISCORD_AUTH_ENABLED = getEnvDefault("DISCORD_AUTH_ENABLED", "false") != "false"
var DISCORD_CLIENT_ID = getEnvCondition("DISCORD_CLIENT_ID", "", DISCORD_AUTH_ENABLED)
var DISCORD_CLIENT_SECRET = getEnvCondition("DISCORD_CLIENT_SECRET", "", DISCORD_AUTH_ENABLED)
var DISCORD_REDIRECT_URL = getEnvDefault("DISCORD_REDIRECT_URL", "https://localhost:"+WEBSERVER_PORT+"/auth/callback")
var DISCORD_AUTH_TOKEN = getEnvDefault("DISCORD_AUTH_TOKEN", randID(32))
var DISCORD_TARGET_GUILD_ID = getEnvCondition("DISCORD_TARGET_GUILD_ID", "", DISCORD_AUTH_ENABLED)

// ---

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

func getEnvDefaultInt(key string, _default string) int64 {
	value := getEnvDefault(key, _default)
	converted, err := strconv.ParseInt(value, 10, 64)

	if err != nil || converted <= 0 {
		log.Fatalf("%s=%s invalid", key, value)
	}

	return converted
}

func getEnvCondition(key, _default string, isMust bool) string {
	if isMust {
		return getEnvMust(key)
	} else {
		return getEnvDefault(key, _default)
	}
}
