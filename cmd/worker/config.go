package main

type (
	Config struct {
		Log `yaml:"log"`

		PG `yaml:"pg"`
	}

	Log struct {
		Level string `env-required:"true" yaml:"log_level" env:"LOG_LEVEL"`
	}

	PG struct {
		PoolMax int    `env-required:"true"  yaml:"pool_max" env:"POOL_MAX"`
		URL     string `env-required:"true"                 env:"PG_URL"`
	}

	CONSUMER struct {
		// из kafka/cfg/config.go
	}
)
