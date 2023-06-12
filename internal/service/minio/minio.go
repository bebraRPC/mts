package minio

import (
	"context"
	"fmt"
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
	DeleteFile(string) error
}

type ClientMinio struct {
	Client     *minio.Client
	BucketName string
	Logger     logger.Interface
}

func NewMinioClient(endpoint, accessKey, secretKey, bucketName string, logger logger.Interface) (*ClientMinio, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	minioClient := &ClientMinio{
		Client:     client,
		BucketName: bucketName,
		Logger:     logger,
	}

	return minioClient, nil
}

func (c *ClientMinio) UploadFile(file io.Reader, filename string) (string, error) {
	contentType := mime.TypeByExtension(filepath.Ext(filename))

	_, err := c.Client.PutObject(context.TODO(), c.BucketName, filename, file, -1, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		c.Logger.Error(fmt.Sprintf("Failed to upload file to MinIO: %v", err))
		return "", fmt.Errorf("failed to upload file to MinIO: %w", err)
	}

	fileURL, _ := c.GetObjectURL(filename)

	return fileURL, nil
}

func (c *ClientMinio) GetFileURL(filename string) (string, error) {
	fileURL, err := c.GetObjectURL(filename)
	if err != nil {
		return "", err
	}
	return fileURL, nil
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
	err := c.Client.RemoveObject(context.TODO(), c.BucketName, filename, minio.RemoveObjectOptions{})
	if err != nil {
		c.Logger.Error(fmt.Sprintf("Failed to delete file from MinIO: %v", err))
		return fmt.Errorf("failed to delete file from MinIO: %w", err)
	}

	return nil
}
