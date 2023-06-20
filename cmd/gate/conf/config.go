package conf

type (
	Config struct {
		HTTP `yaml:"http"`
		Log  `yaml:"log"`

		PG `yaml:"pg"`
	}

	HTTP struct {
		Port string `env-required:"true" yaml:"port" env:"HTTP_PORT"`
	}

	Log struct {
		Level string `env-required:"true" yaml:"log_level" env:"LOG_LEVEL"`
	}

	PG struct {
		PoolMax int    `env-required:"true"  yaml:"pool_max" env:"POOL_MAX"`
		URL     string `env-required:"true"                 env:"PG_URL"`
	}
)
