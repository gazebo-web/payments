package client

import (
	"context"
	"gitlab.com/ignitionrobotics/billing/payments/pkg/api"
)

// client contains the HTTP client to connect to the payments API.
type client struct{}

// CreateSession performs an HTTP request to create a payment session in the Payments API.
func (c *client) CreateSession(ctx context.Context, req api.CreateSessionRequest) (api.CreateSessionResponse, error) {
	panic("implement me")
}

// ListInvoices performs an HTTP request to list all the available invoices of a certain user.
func (c *client) ListInvoices(ctx context.Context, req api.ListInvoicesRequest) (api.ListInvoicesResponse, error) {
	panic("implement me")
}

// Client holds methods to interact with a api.PaymentsV1 service.
type Client interface {
	api.PaymentsV1
}

// NewClient initializes a new api.PaymentsV1 client implementation using an HTTP client.
func NewClient() Client {
	return &client{}
}
