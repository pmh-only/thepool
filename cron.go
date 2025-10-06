package main

import (
	"log"
	"strconv"
	"time"
)

func startPurgeJobLoop() {
	maxPoolSizeRaw := getEnvDefault("CRONJOB_POOL_SIZE_LIMIT_MB", "1024")

	maxPoolSizeMB, err := strconv.ParseInt(maxPoolSizeRaw, 10, 64)
	if err != nil || maxPoolSizeMB <= 0 {
		log.Fatalf("CRONJOB_POOL_SIZE_LIMIT_MB=%q invalid", maxPoolSizeRaw)
	}

	for {
		chunks := purgeChunk(maxPoolSizeMB)
		deleteChunk(chunks)

		time.Sleep(30 * time.Second)
	}
}
