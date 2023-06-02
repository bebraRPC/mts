package adapter

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"log"
	"mime/multipart"
	"mts/internal/gate/domain"
)

type MinioStorage struct {
	client *minio.Client
	bucket string
}

func NewMinioStorage(endpoint, accessKey, secretKey, bucket string) (*MinioStorage, error) {
	ctx := context.Background()
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalln("failed to create Minio client: %w", err)
		return nil, err
	}

	err = minioClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
	if err != nil {
		exists, errBucketExists := minioClient.BucketExists(ctx, bucket)
		if errBucketExists == nil && exists {
			return nil, fmt.Errorf("failed to create Minio bucket: %w", err)
		} else {
			log.Fatalln(err)
			return nil, err
		}
	}

	return &MinioStorage{client: minioClient, bucket: bucket}, nil
}

func (m *MinioStorage) SaveImage(image domain.Image, file multipart.File) error {
	objectName := image.ID
	_, err := m.client.PutObject(context.Background(), m.bucket, objectName, file, -1, minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to save image to Minio: %w", err)
	}

	return nil
}

func (m *MinioStorage) GetImageByID(id string) (domain.Image, error) {
	objectName := id
	object, err := m.client.GetObject(context.Background(), m.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return domain.Image{}, fmt.Errorf("failed to get image from Minio: %w", err)
	}

	return domain.Image{
		ID: object.
	}
}
