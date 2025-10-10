package main

import (
	"time"
)

func startPurgeJobLoop() {
	for {
		chunks := purgeChunk(CRONJOB_POOL_SIZE_LIMIT_MB * MB)
		deleteChunks(chunks)

		time.Sleep(30 * time.Second)
	}
}
