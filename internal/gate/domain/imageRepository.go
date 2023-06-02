package domain

type ImageRepository interface {
	SaveImage(image Image) error
	GetImageByID(id string) (Image, error)
}
