package kafka

type ImgKafka struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	OriginalURL string `json:"originalURL"`
}

// для consumer, для чтения, поиска в бд, заполнения 512/256/16 и originalURL для загрузки из minio

//func (i *ImgKafka) FromKafka()

//func (i *ImgKafka) ToKafka()

func FromKafka() {

}
