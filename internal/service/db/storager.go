package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/menyasosali/mts/internal/domain"
	"github.com/menyasosali/mts/pkg/logger"
	"github.com/menyasosali/mts/pkg/postgres"
)

type StoreInterface interface {
	SaveImage(domain.ImgDescriptor) error
	GetImageByID(string) (domain.ImgDescriptor, error)
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

func (s *Store) SaveImage(image domain.ImgDescriptor) error {
	query := `
		INSERT INTO images (image_id, name, origin_url, jpg512_url, jpg256_url, jpg16_base64_url)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (image_id) DO UPDATE
		SET name = $2, origin_url = $3, jpg512_url = $4, jpg256_url = $5, jpg16_base64_url = $6
	`

	_, err := s.Pg.Pool.Exec(s.Ctx, query, image.ID, image.Name, image.URL, image.URL512, image.URL256, image.URL16)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("Failed to save image in database: %v", err))
		return fmt.Errorf("failed to save image in database: %w", err)
	}

	return nil
}

func (s *Store) GetImageByID(imageID string) (*domain.ImgDescriptor, error) {
	query := `
		SELECT image_id, name, origin_url, jpg512_url, jpg256_url, jpg16_base64_url
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
