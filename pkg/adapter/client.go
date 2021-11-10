package adapter

import (
	customers "gitlab.com/ignitionrobotics/billing/customers/pkg/api"
	"gitlab.com/ignitionrobotics/billing/payments/pkg/api"
)

// Client wraps a payment service client such as Stripe to be used as an adapter.
type Client interface {
	// CreateCustomer creates a customer in the context of the payment service.
	CreateCustomer(application, handle string) (string, error)

	// CreateSession creates a session in the context of the payment service. It's usually
	// used to create new checkout sessions.
	CreateSession(req api.CreateSessionRequest, cus customers.CustomerResponse) (api.CreateSessionResponse, error)

	// GenerateChargeRequest generates an api.ChargeRequest out from the given body and a set
	// of parameters
	GenerateChargeRequest(body []byte, params map[string][]string) (api.ChargeRequest, error)
}
