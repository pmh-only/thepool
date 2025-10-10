package main

import (
	"context"
	"crypto/tls"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var minioClient *minio.Client

func createMinioConnection() {
	var transport http.RoundTripper = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		DisableCompression: true,
	}

	var err error
	minioClient, err = minio.New(MINIO_ENDPOINT, &minio.Options{
		Creds: credentials.NewStaticV4(
			MINIO_ACCESS_KEY_ID,
			MINIO_SECRET_ACCESS_KEY,
			""),
		Secure:    MINIO_ENDPOINT_SECURE,
		Transport: transport,
	})

	if err != nil {
		log.Fatal(err)
	}
}

func uploadChunk(chunkId string, chunkSize int64, reader io.Reader) error {
	_, err := minioClient.PutObject(
		context.Background(),
		MINIO_BUCKET_NAME,
		MINIO_BUCKET_KEY_PREFIX+chunkId,
		reader,
		chunkSize,
		minio.PutObjectOptions{
			ContentType: "application/octet-stream",
		},
	)

	return err
}

func deleteChunks(chunkIds []string) {
	for _, chunkId := range chunkIds {
		minioClient.RemoveObject(
			context.Background(),
			MINIO_BUCKET_NAME,
			MINIO_BUCKET_KEY_PREFIX+chunkId,
			minio.RemoveObjectOptions{
				ForceDelete:      true,
				GovernanceBypass: true,
			},
		)
	}
}
