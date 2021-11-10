package client

import (
	"context"
	"gitlab.com/ignitionrobotics/billing/payments/pkg/api"
	"gitlab.com/ignitionrobotics/web/ign-go/encoders"
	"gitlab.com/ignitionrobotics/web/ign-go/net"
	"net/http"
	"net/url"
	"time"
)

// client contains the HTTP client to connect to the payments API.
type client struct {
	client net.Client
}

// CreateSession performs an HTTP request to create a payment session in the Payments API.
func (c *client) CreateSession(ctx context.Context, in api.CreateSessionRequest) (api.CreateSessionResponse, error) {
	var out api.CreateSessionResponse
	if err := c.client.Call(ctx, "CreateSession", &in, &out); err != nil {
		return api.CreateSessionResponse{}, err
	}
	return out, nil
}

// ListInvoices performs an HTTP request to list all the available invoices of a certain user.
func (c *client) ListInvoices(ctx context.Context, req api.ListInvoicesRequest) (api.ListInvoicesResponse, error) {
	panic("implement me")
}

// Client holds methods to interact with a api.PaymentsV1 service.
type Client interface {
	api.PaymentsV1
}

// NewPaymentsClientV1 initializes a new api.PaymentsV1 client implementation using an HTTP client.
func NewPaymentsClientV1(baseURL *url.URL, timeout time.Duration) Client {
	endpoints := map[string]net.EndpointHTTP{
		"CreateSession": {
			Method: http.MethodPost,
			Path:   "/payments/session",
		},
	}
	return &client{
		client: net.NewClient(net.NewCallerHTTP(baseURL, endpoints, timeout), encoders.JSON),
	}
}
