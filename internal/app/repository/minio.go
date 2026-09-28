package repository

import (
	"context"
	"mime/multipart"

	"github.com/minio/minio-go/v7"
)

func (r *Repository) UploadFile(ctx context.Context, key string, header *multipart.FileHeader, contentType string) error {
	file, err := header.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = r.minio.PutObject(ctx, r.minioBucket, key, file, header.Size, minio.PutObjectOptions{ContentType: contentType})
	return err
}

func (r *Repository) RemoveFile(ctx context.Context, key string) error {
	return r.minio.RemoveObject(ctx, r.minioBucket, key, minio.RemoveObjectOptions{})
}
