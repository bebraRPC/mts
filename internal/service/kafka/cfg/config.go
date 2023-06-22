package kafkacfg

type ConsumerConfig struct {
	Brokers []string `env-required:"true" yaml:"brokers" env:"KAFKA_BROKERS"`
	Topic   string   `env-required:"true" yaml:"topic" env:"KAFKA_TOPIC"`
	GroupID string   `env-required:"true" yaml:"group_id" env:"KAFKA_GROUP_ID"`
}

type ProducerConfig struct {
	Brokers []string `env-required:"true" yaml:"brokers" env:"KAFKA_BROKERS"`
	Topic   string   `env-required:"true" yaml:"topic" env:"KAFKA_TOPIC"`
}
