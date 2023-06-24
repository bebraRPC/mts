package config

type GateConfig struct {
	Log      LogConfig      `yaml:"logger"`
	Postgres PostgresConfig `yaml:"postgres"`
	Minio    MinioConfig    `yaml:"minio"`
	Kafka    KafkaConfig    `yaml:"kafka"`
	HTTP     HTTPConfig     `yaml:"http"`
}

type WorkerConfig struct {
	Log   LogConfig   `yaml:"logger"`
	Kafka KafkaConfig `yaml:"kafka"`
	Minio MinioConfig `yaml:"minio"`
}

type PostgresConfig struct {
	PoolMax int    `env-required:"true" yaml:"pool_max" env:"PG_POOL_MAX" env-default:"2"`
	URL     string `env-required:"true"                 env:"PG_URL" env-default:"postgres://user:pass@localhost:5432/postgres"`
}

type MinioConfig struct {
	Endpoint   string `yaml:"endpoint" env:"" env-default:"localhost:9000"`
	AccessKey  string `yaml:"access_key" env:"" env-default:"accesskey"`
	SecretKey  string `yaml:"secret_key" env:"" env-default:"secretkey"`
	BucketName string `yaml:"bucket_name" env:"" env-default:"mts"`
}

type LogConfig struct {
	Level string `env-required:"true" yaml:"log_level"   env:"LOG_LEVEL" env-default:"debug"`
}

type KafkaConfig struct {
	Brokers []string `env-required:"true" yaml:"brokers" env:"KAFKA_BROKERS" env-default:"localhost:9092"`
	Topic   string   `env-required:"true" yaml:"topic" env:"KAFKA_TOPIC" env-default:"image-topic"`
	GroupID string   `env-required:"true" yaml:"group_id" env:"KAFKA_GROUP_ID" env-default:"worker-group"`
}

type HTTPConfig struct {
	Port string `env-required:"true" yaml:"port" env:"HTTP_PORT" env-default:"8080"`
}
