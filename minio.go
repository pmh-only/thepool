package main

import (
	"context"
	"io"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var minioClient *minio.Client

func createMinioConnection() {
	endpoint := getEnvMust("MINIO_ENDPOINT")

	accessKeyID := getEnvMust("MINIO_ACCESS_KEY_ID")
	secretAccessKey := getEnvMust("MINIO_SECRET_ACCESS_KEY")

	var err error
	minioClient, err = minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: false,
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
		chunkSizeMB,
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
