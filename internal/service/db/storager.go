package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/menyasosali/mts/internal/domain"
	"github.com/menyasosali/mts/pkg/logger"
	"github.com/menyasosali/mts/pkg/postgres"
)

type Storager interface {
	SaveImage(image domain.Image) error
	GetImageByID(id string) (domain.Image, error)
}

type Store struct {
	pg     *postgres.Postgres
	Logger logger.Interface
}

func NewStore(pg *postgres.Postgres, logger logger.Interface) *Store {
	return &Store{pg: pg, Logger: logger}
}

func (s *Store) SaveImage(image domain.Image) error {
	query := `
		INSERT INTO images (image_id, origin_url, jpg512_url, jpg256_url, jpg16_base64_url)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (image_id) DO UPDATE
		SET origin_url = $2, jpg512_url = $3, jpg256_url = $4, jpg16_base64_url = $5
	`

	_, err := s.pg.Pool.Exec(context.Background(), query, image.ID, image.URL, image.URL512, image.URL256, image.URL16)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("Failed to save image in database: %v", err))
		return fmt.Errorf("failed to save image in database: %w", err)
	}

	return nil
}

func (s *Store) GetImageByID(imageID string) (*domain.Image, error) {
	query := `
		SELECT image_id, origin_url, jpg512_url, jpg256_url, jpg16_base64_url
		FROM images
		WHERE image_id = $1
	`

	row := s.pg.Pool.QueryRow(context.Background(), query, imageID)

	image := &domain.Image{}
	err := row.Scan(&image.ID, &image.URL, &image.URL512, &image.URL256, &image.URL16)
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
