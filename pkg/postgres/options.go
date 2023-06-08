package postgres

import "time"

type Option func(*Postgres)

func MaxPoolSize(size int) Option {
	return func(postgres *Postgres) {
		postgres.maxPoolSize = size
	}
}

func ConnAttempts(attempts int) Option {
	return func(postgres *Postgres) {
		postgres.connAttempts = attempts
	}
}

func ConnTimeout(timeout time.Duration) Option {
	return func(postgres *Postgres) {
		postgres.connTimeout = timeout
	}
}
