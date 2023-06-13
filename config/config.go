package config

type (
	Config struct {
		HTTP  `yaml:"http"`
		Log   `yaml:"log"`
		MINIO `yaml:"minio"`
		PG    `yaml:"pg"`
		Kafka `yaml:"kafka"`
	}

	HTTP struct {
		Port string `env-required:"true" yaml:"port" env:"HTTP_PORT"`
	}

	Log struct {
		Level string `env-required:"true" yaml:"log_level" env:"LOG_LEVEL"`
	}

	MINIO struct {
		Port string `env-required:"true" yaml:"port" emv:"MINIO_PORT"`
	}

	PG struct {
		PoolMax int `env-required:"true"  yaml:"pool_max" env:"POOL_MAX"`
	}

	Kafka struct {
		Brokers []string `env-required:"true" yaml:"brokers" env:"KAFKA_BROKERS"`
		Topic   string   `env-required:"true" yaml:"topic" env:"KAFKA_TOPIC"`
		GroupID string   `env-required:"true" yaml:"group_id" env:"KAFKA_GROUP_ID"`
	}
)
