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
	endpoint := getEnvMust("MINIO_ENDPOINT")

	accessKeyID := getEnvMust("MINIO_ACCESS_KEY_ID")
	secretAccessKey := getEnvMust("MINIO_SECRET_ACCESS_KEY")

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
	minioClient, err = minio.New(endpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure:    getEnvDefault("MINIO_ENDPOINT_SECURE", "true") != "false",
		Transport: transport,
	})

	if err != nil {
		log.Fatalln(err)
	}
}

func uploadChunk(chunkId string, chunkSizeMB int64, reader io.Reader) error {
	_, err := minioClient.PutObject(
		context.Background(),
		getEnvDefault("MINIO_BUCKET_NAME", "mybucket"),
		getEnvDefault("MINIO_BUCKET_KEY_PREFIX", "")+chunkId,
		reader,
		chunkSizeMB*1024*1024,
		minio.PutObjectOptions{
			ContentType: "application/octet-stream",
		},
	)

	return err
}

func deleteChunk(chunkIds []string) {
	for _, chunkId := range chunkIds {
		minioClient.RemoveObject(
			context.Background(),
			getEnvDefault("MINIO_BUCKET_NAME", "mybucket"),
			getEnvDefault("MINIO_BUCKET_KEY_PREFIX", "")+chunkId,
			minio.RemoveObjectOptions{
				ForceDelete:      true,
				GovernanceBypass: true,
			},
		)
	}
}
