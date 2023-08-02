package minio

import (
	"bytes"
	"context"
	"fmt"
	"github.com/menyasosali/mts/config"
	"github.com/menyasosali/mts/pkg/logger"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"mime"
	"net/url"
	"path/filepath"
)

// interface для minio

type InterfaceMinio interface {
	UploadFile(context.Context, io.Reader, string) (string, error)
	GetFileURL(context.Context, string) (string, error)
	DownloadFile(context.Context, string) ([]byte, error)
	DeleteFile(context.Context, string) error
}

type ClientMinio struct {
	Logger     logger.Interface
	Client     *minio.Client
	BucketName string
}

func NewMinioClient(logger logger.Interface, cfg config.MinioConfig) (*ClientMinio, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: false,
	})
	logger.Info(fmt.Sprintf("AccessKey: %s\n SecretKey: %s\n", cfg.AccessKey, cfg.SecretKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	minioClient := &ClientMinio{
		Logger:     logger,
		Client:     client,
		BucketName: cfg.BucketName,
	}

	return minioClient, nil
}

func (c *ClientMinio) UploadFile(ctx context.Context, file []byte, filename string) (string, error) {
	contentType := mime.TypeByExtension(filepath.Ext(filename))
	location := "serv"

	err := c.Client.MakeBucket(ctx, c.BucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := c.Client.BucketExists(ctx, c.BucketName)
		if errBucketExists == nil && exists {
			c.Logger.Info(fmt.Sprintf("We already own %s\n", c.BucketName))
		} else {
			c.Logger.Error(err)
		}
	} else {
		c.Logger.Info(fmt.Sprintf("Successfully created %s\n", c.BucketName))
	}

	_, err = c.Client.PutObject(ctx, c.BucketName, filename, bytes.NewReader(file), -1, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		c.Logger.Error(fmt.Sprintf("Failed to upload file to MinIO: %v", err))
		return "", fmt.Errorf("failed to upload file to MinIO: %w", err)
	}

	fileURL, _ := c.GetObjectURL(ctx, filename)

	return fileURL, nil
}

func (c *ClientMinio) DownloadFile(ctx context.Context, filename string) ([]byte, error) {
	object, err := c.Client.GetObject(ctx, c.BucketName, filename, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to download image from MinIO: %w", err)
	}
	defer object.Close()

	data, err := io.ReadAll(object)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}

	return data, nil
}

func (c *ClientMinio) GetObjectURL(ctx context.Context, filename string) (string, error) {
	baseURL := c.Client.EndpointURL()

	filePath := fmt.Sprintf("/%s/%s", c.BucketName, filename)
	objectURL := baseURL.ResolveReference(&url.URL{Path: filePath}).String()

	if objectURL == "" {
		c.Logger.Error(fmt.Sprintf("Failed to get object url from MinIO"))
		return "", fmt.Errorf("failed to get object url from MinIO")
	}

	return objectURL, nil
}

func (c *ClientMinio) DeleteFile(ctx context.Context, filename string) error {
	err := c.Client.RemoveObject(ctx, c.BucketName, filename, minio.RemoveObjectOptions{})
	if err != nil {
		c.Logger.Error(fmt.Sprintf("Failed to delete file from MinIO: %v", err))
		return fmt.Errorf("failed to delete file from MinIO: %w", err)
	}

	return nil
}
