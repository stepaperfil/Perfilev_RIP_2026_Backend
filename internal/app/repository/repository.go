package repository

import (
	"github.com/minio/minio-go/v7"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository struct {
	db          *gorm.DB
	minio       *minio.Client
	minioBucket string
}

func NewRepository(dsn string, minioClient *minio.Client, minioBucket string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &Repository{db: db, minio: minioClient, minioBucket: minioBucket}, nil
}
