package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/menyasosali/mts/internal/domain"
	"github.com/menyasosali/mts/pkg/logger"
	"github.com/menyasosali/mts/pkg/postgres"
)

type StoreInterface interface {
	UploadImage(string, string) (string, error)
	GetImageByID(string) (*domain.ImgDescriptor, error)
	UpdateImage(descriptor domain.ImgDescriptor) error
}

type Store struct {
	Ctx    context.Context
	Logger logger.Interface
	Pg     *postgres.Postgres
}

func NewStore(ctx context.Context, logger logger.Interface, pg *postgres.Postgres) *Store {
	return &Store{
		Ctx:    ctx,
		Logger: logger,
		Pg:     pg,
	}
}

func (s *Store) UploadImage(name, originalURL string) (string, error) {
	image := domain.ImgDescriptor{
		Name: name,
		URL:  originalURL,
	}
	query := `
		INSERT INTO images (image_id, name, original_url, url_512, url_256, url_16)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (image_id) DO UPDATE
		SET name = $2, original_url = $3, url_512 = $4, url_256 = $5, url_16 = $6
		RETURNING image_id
	`

	image.ID = uuid.New().String()
	err := s.Pg.Pool.QueryRow(s.Ctx, query, image.ID, image.Name, image.URL, image.URL512, image.URL256, image.URL16).Scan(&image.ID)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("Failed to save image in database: %v", err))
		return "", fmt.Errorf("failed to save image in database: %w", err)
	}

	return image.ID, nil
}

func (s *Store) GetImageByID(imageID string) (*domain.ImgDescriptor, error) {
	query := `
		SELECT image_id, name, original_url, url_512, url_256, url_16
		FROM images
		WHERE image_id = $1
	`

	row := s.Pg.Pool.QueryRow(s.Ctx, query, imageID)

	image := &domain.ImgDescriptor{}
	err := row.Scan(&image.ID, &image.Name, &image.URL, &image.URL512, &image.URL256, &image.URL16)
	if err != nil {
		if err == sql.ErrNoRows {
			s.Logger.Error(fmt.Sprintf("Image not found in database: %v", err))
			return nil, fmt.Errorf("image not found in database: %v", err)
		}
		s.Logger.Error(fmt.Sprintf("Failed to get image from database: %v", err))
		return nil, fmt.Errorf("failed to get image from database: %w", err)
	}

	return image, nil
}

func (s *Store) UpdateImage(img domain.ImgDescriptor) error {
	query := `
		UPDATE images
		SET url_512 = $2, url_256 = $3, url_16 = $4
		WHERE image_id = $1
	`

	_, err := s.Pg.Pool.Exec(s.Ctx, query, img.ID, img.URL512, img.URL256, img.URL16)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("Failed to update image in database: %v", err))
		return fmt.Errorf("failed to update image in database: %w", err)
	}

	s.Logger.Info(fmt.Sprintf("The image was successfully updated"))
	return nil
}
