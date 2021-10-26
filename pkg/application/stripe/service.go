package stripe

import (
	"context"
	credits "gitlab.com/ignitionrobotics/billing/credits/pkg/api"
	"gitlab.com/ignitionrobotics/billing/payments/pkg/api"
	"io"
	"log"
)

// service contains the business logic to manage Stripe payments.
type service struct {
	logger  *log.Logger
	credits credits.CreditsV1
}

// Charge charges a certain amount of money to a given user.
func (s *service) Charge(ctx context.Context, req api.ChargeRequest) (api.ChargeResponse, error) {
	panic("implement me")
}

// CreateSession creates a session for a user to pay for a certain product or service.
// This token is intended to allow external interfaces to interact with the payment provider on behalf of the user.
func (s *service) CreateSession(ctx context.Context, req api.CreateSessionRequest) (api.CreateSessionResponse, error) {
	panic("implement me")
}

// ListInvoices returns a list of invoices of the given user.
func (s *service) ListInvoices(ctx context.Context, req api.ListInvoicesRequest) (api.ListInvoicesResponse, error) {
	panic("implement me")
}

// Service holds methods to interact with different payments systems.
type Service interface {
	api.ChargerV1
	api.PaymentsV1
}

// NewService initializes a new Service implementation using Stripe.
func NewService(credits credits.CreditsV1, logger *log.Logger) Service {
	if logger == nil {
		logger = log.New(io.Discard, "", log.LstdFlags)
	}
	return &service{
		logger:  logger,
		credits: credits,
	}
}
