package application

import (
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/client"
	"gitlab.com/ignitionrobotics/billing/payments/internal/conf"
)

// NewStripeClient initializes a new Stripe client using the provided conf.Stripe config.
func NewStripeClient(cfg conf.Stripe) *client.API {
	var backendURL *string
	if len(cfg.URL) > 0 {
		backendURL = &cfg.URL
	}
	config := stripe.BackendConfig{
		URL: backendURL,
	}
	return client.New(cfg.SecretKey, &stripe.Backends{
		API:     stripe.GetBackendWithConfig(stripe.APIBackend, &config),
		Connect: stripe.GetBackendWithConfig(stripe.ConnectBackend, &config),
		Uploads: stripe.GetBackendWithConfig(stripe.UploadsBackend, &config),
	})
}
