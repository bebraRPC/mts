package minio

import (
	"bytes"
	"context"
	"fmt"
	miniocfg "github.com/menyasosali/mts/internal/service/minio/cfg"
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
	UploadFile(io.Reader, string) (string, error)
	GetFileURL(string) (string, error)
	DownloadFile(string) ([]byte, error)
	DeleteFile(string) error
}

type ClientMinio struct {
	Ctx        context.Context
	Logger     logger.Interface
	Client     *minio.Client
	BucketName string
}

func NewMinioClient(ctx context.Context, logger logger.Interface, cfg miniocfg.Config) (*ClientMinio, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	minioClient := &ClientMinio{
		Ctx:        ctx,
		Logger:     logger,
		Client:     client,
		BucketName: cfg.BucketName,
	}

	return minioClient, nil
}

func (c *ClientMinio) UploadFile(file []byte, filename string) (string, error) {
	contentType := mime.TypeByExtension(filepath.Ext(filename))

	_, err := c.Client.PutObject(c.Ctx, c.BucketName, filename, bytes.NewReader(file), -1, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		c.Logger.Error(fmt.Sprintf("Failed to upload file to MinIO: %v", err))
		return "", fmt.Errorf("failed to upload file to MinIO: %w", err)
	}

	fileURL, _ := c.GetObjectURL(filename)

	return fileURL, nil
}

func (c *ClientMinio) DownloadFile(filename string) ([]byte, error) {
	object, err := c.Client.GetObject(c.Ctx, c.BucketName, filename, minio.GetObjectOptions{})
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

func (c *ClientMinio) GetObjectURL(filename string) (string, error) {
	baseURL := c.Client.EndpointURL()

	filePath := fmt.Sprintf("/%s/%s", c.BucketName, filename)
	objectURL := baseURL.ResolveReference(&url.URL{Path: filePath}).String()

	if objectURL == "" {
		c.Logger.Error(fmt.Sprintf("Failed to get object url from MinIO"))
		return "", fmt.Errorf("failed to get object url from MinIO")
	}

	return objectURL, nil
}

func (c *ClientMinio) DeleteFile(filename string) error {
	err := c.Client.RemoveObject(c.Ctx, c.BucketName, filename, minio.RemoveObjectOptions{})
	if err != nil {
		c.Logger.Error(fmt.Sprintf("Failed to delete file from MinIO: %v", err))
		return fmt.Errorf("failed to delete file from MinIO: %w", err)
	}

	return nil
}
