package conf

import (
	"github.com/caarlos0/env/v6"
	"time"
)

// Config contains the needed config to start the Payments HTTP server.
type Config struct {
	Port             uint          `env:"PAYMENTS_HTTP_SERVER_PORT" envDefault:"80"`
	StripeSigningKey string        `env:"PAYMENTS_STRIPE_SIGNING_KEY,required"`
	Timeout          time.Duration `env:"PAYMENTS_CIRCUIT_BREAKER_TIMEOUT" envDefault:"30s"`
}

// Parse fills Config data from an external source.
func (c *Config) Parse() error {
	return env.Parse(c)
}
