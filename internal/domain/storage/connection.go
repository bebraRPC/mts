package storage

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/menyasosali/mts/internal/domain"
)

// orm

// интерфейфс
// 3 func: create update delete для интерфейса в domain storage

type StoreInterface interface {
	CreateImg(domain.ImgDescriptor) error
	UpdateImg(domain.ImgDescriptor) error
	DeleteImg(domain.ImgDescriptor) error
}

type Connection struct {
	DB *gorm.DB
}

func NewConnection(connectionString string) (*Connection, error) {
	db, err := gorm.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %w", err)
	}

	return &Connection{
		DB: db,
	}, nil
}
