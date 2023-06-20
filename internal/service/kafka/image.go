package kafka

type ImgKafka struct {
	ID          string //uuid
	Name        string
	OriginalURL string
}

// для consumer, для чтения, поиска в бд, заполнения 512/256/16 и originalURL для загрузки из minio
