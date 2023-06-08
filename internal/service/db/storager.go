package db

import "mts/internal/domain"

type Storager interface {
	SaveImage(image domain.Image) error
	GetImageByID(id string) (domain.Image, error)
}

type Store struct {
}

func NewStore() *Store {
	return &Store{}
}
