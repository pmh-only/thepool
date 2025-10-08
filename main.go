package main

import _ "github.com/joho/godotenv/autoload"

func main() {
	createMysqlConnection()
	createMinioConnection()

	go startPurgeJobLoop()

	setupMux()
	openWebserver()
}
