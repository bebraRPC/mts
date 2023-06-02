package service

import "mts/internal/gate/domain"

type ImageService struct {
	repo domain.ImageRepository
}

func NewImageService(repo domain.ImageRepository) *ImageService {
	return &ImageService{
		repo: repo,
	}
}

func (s *ImageService) Do() {

}
