package storage

import (
	"context"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/menyasosali/mts/internal/domain"
	"github.com/menyasosali/mts/pkg/logger"
)

// создаю entity которые залетают в бд postgres

// imageupload()
// image get (

// реализ интерфейс из connect
// функции принимает на вход структуру Image и возвр

type Storage struct {
	Ctx        context.Context
	Logger     logger.Interface
	Connection *Connection
}

func NewStorage(ctx context.Context, logger logger.Interface, connection *Connection) *Storage {
	return &Storage{
		Ctx:        ctx,
		Logger:     logger,
		Connection: connection,
	}
}

func (s *Storage) CreateImg(img *domain.ImgDescriptor) error {
	err := s.Connection.DB.Create(&img).Error
	if err != nil {
		errMsg := fmt.Errorf("failed to create image: %w", err)
		s.Logger.Error(errMsg)
		return errMsg
	}

	return nil
}

func (s *Storage) UpdateImg(img *domain.ImgDescriptor) error {
	err := s.Connection.DB.Save(&img).Error
	if err != nil {
		errMsg := fmt.Errorf("failed to update image: %w", err)
		s.Logger.Error(errMsg)
		return errMsg
	}

	return nil
}

func (s *Storage) DeleteImg(img *domain.ImgDescriptor) error {
	err := s.Connection.DB.Delete(&img.ID).Error
	if err != nil {
		errMsg := fmt.Errorf("failed to delete image: %w", err)
		s.Logger.Error(errMsg)
		return errMsg
	}

	return nil
}

func (s *Storage) ImageUpload(img *domain.ImgDescriptor) error {
	err := s.CreateImg(img)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) GetImageByID(id string) (*domain.ImgDescriptor, error) {
	img := &domain.ImgDescriptor{ID: id}
	err := s.Connection.DB.First(&img).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		errMsg := fmt.Errorf("failed to get image by ID: %w", err)
		s.Logger.Error(errMsg)
		return nil, errMsg
	}

	return img, nil
}
