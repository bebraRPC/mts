package producerCfg

type Config struct {
	Producer `yaml:"producer"`
}

type Producer struct {
	Brokers []string `env-required:"true" yaml:"brokers" env:"KAFKA_BROKERS"`
	Topic   string   `env-required:"true" yaml:"topic" env:"KAFKA_TOPIC"`
}
