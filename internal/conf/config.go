package conf

import (
	"github.com/caarlos0/env/v6"
	"time"
)

// Stripe contains the needed config to interact with the stripe API.
type Stripe struct {
	// SigningKey is the key used when checking webhook event signatures.
	SigningKey string `env:"PAYMENTS_STRIPE_SIGNING_KEY,required"`

	// SecretKey is the key used to allow the stripe client use the stripe API.
	SecretKey string `env:"PAYMENTS_STRIPE_SECRET_KEY,required"`

	// URL is the backend stripe API url, only used for testing purposes.
	URL string `env:"PAYMENTS_STRIPE_URL"`
}

// Parse fills Stripe data from an external source.
func (c *Stripe) Parse() error {
	return env.Parse(c)
}

// Config contains the needed config to start the Payments HTTP server.
type Config struct {
	// Stripe contains configuration for the stripe client.
	Stripe Stripe

	// Port is the TCP port to listen to for incoming HTTP requests.
	Port uint `env:"PAYMENTS_HTTP_SERVER_PORT" envDefault:"80"`

	// Timeout is used as the amount of time requests originated from the payments service should wait until it fails due
	// to timeout.
	Timeout time.Duration `env:"PAYMENTS_CIRCUIT_BREAKER_TIMEOUT" envDefault:"30s"`

	// CreditsURL contains the URL to the credits service.
	CreditsURL string `env:"PAYMENTS_CREDITS_SERVICE_URL,required"`
}

// Parse fills Config data from an external source.
func (c *Config) Parse() error {
	if err := c.Stripe.Parse(); err != nil {
		return err
	}
	return env.Parse(c)
}
