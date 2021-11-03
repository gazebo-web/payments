package application

import (
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/client"
)

// NewStripeClient initializes a new Stripe client that will use the given key to connect to the given URL.
// 	The URL is usually provided when testing locally (using stripe-mock), the default option (nil) points to the
// 	production Stripe API.
// 	If you want to run tests on the production Stripe API, use the testing key provided in the Stripe Dashboard,
// 	and keep the URL set to nil.
func NewStripeClient(key string, url *string) *client.API {
	config := stripe.BackendConfig{
		URL: url,
	}
	return client.New(key, &stripe.Backends{
		API:     stripe.GetBackendWithConfig(stripe.APIBackend, &config),
		Connect: stripe.GetBackendWithConfig(stripe.ConnectBackend, &config),
		Uploads: stripe.GetBackendWithConfig(stripe.UploadsBackend, &config),
	})
}
