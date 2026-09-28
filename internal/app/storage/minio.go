package storage

import (
	"fmt"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func Endpoint() string {
	return getEnv("MINIO_ENDPOINT", "localhost:9000")
}

func Bucket() string {
	return getEnv("MINIO_BUCKET", "tco-assets")
}

func NewClient() (*minio.Client, error) {
	return minio.New(Endpoint(), &minio.Options{
		Creds:  credentials.NewStaticV4(getEnv("MINIO_ACCESS_KEY", "minioadmin"), getEnv("MINIO_SECRET_KEY", "minioadmin"), ""),
		Secure: false,
	})
}

func ObjectURL(key string) string {
	if key == "" {
		return ""
	}
	return fmt.Sprintf("http://%s/%s/%s", Endpoint(), Bucket(), key)
}
