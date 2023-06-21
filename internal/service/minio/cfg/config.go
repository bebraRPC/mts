package miniocfg

type Config struct {
	MINIO `yaml:"minio"`
}

type MINIO struct {
	Port       string `env-required:"true" yaml:"port" emv:"MINIO_PORT"`
	Endpoint   string `env-required:"true" yaml:"endpoint" env:"MINIO_ENDPOINT"`
	AccessKey  string `env-required:"true" yaml:"accessKey" env:"MINIO_ACCESS_KEY"`
	SecretKey  string `env-required:"true" yaml:"secretKey" env:"MINIO_SECRET_KEY"`
	BucketName string `env-required:"true" yaml:"bucketName" env:"MINIO_BUCKET_NAME"`
}
