package cfg

type Config struct {
	MINIO `yaml:"minio"`
}

type MINIO struct {
	Port string `env-required:"true" yaml:"port" emv:"MINIO_PORT"`
}
