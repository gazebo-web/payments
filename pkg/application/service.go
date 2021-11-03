package application

import (
	"context"
	"github.com/stripe/stripe-go/v72/client"
	credits "gitlab.com/ignitionrobotics/billing/credits/pkg/api"
	customers "gitlab.com/ignitionrobotics/billing/customers/pkg/api"
	"gitlab.com/ignitionrobotics/billing/payments/pkg/api"
	"io"
	"log"
	"time"
)

// service contains the business logic to manage payments on different billing systems such as Stripe.
type service struct {
	logger    *log.Logger
	credits   credits.CreditsV1
	customers customers.CustomersV1
	stripe    *client.API
	timeout   time.Duration
}

// Charge charges a certain amount of money to a given user.
func (s *service) Charge(ctx context.Context, req api.ChargeRequest) (api.ChargeResponse, error) {
	s.logger.Printf("Processing charge request: %+v\n", req)

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Main thread
	ch := make(chan api.ChargeResponse, 1)
	errs := make(chan error, 1)
	go func() {
		customerResponse, err := s.customers.GetCustomerByID(ctx, customers.GetCustomerByIDRequest{
			ID:          req.Customer,
			Service:     string(req.Service),
			Application: req.Application,
		})
		if err != nil {
			errs <- err
			return
		}

		_, err = s.credits.IncreaseCredits(ctx, credits.IncreaseCreditsRequest{
			Transaction: credits.Transaction{
				Handle:      customerResponse.Handle,
				Amount:      req.Amount,
				Currency:    req.Currency,
				Application: req.Application,
			},
		})
		if err != nil {
			errs <- err
			return
		}

		ch <- api.ChargeResponse{}
	}()

	select {
	case <-ctx.Done(): // Circuit breaker
		s.logger.Println("Context error:", ctx.Err())
		return api.ChargeResponse{}, ctx.Err()
	case err := <-errs: // Error handler
		s.logger.Println("Failed to process charge:", err)
		return api.ChargeResponse{}, err
	case res := <-ch: // Post-processing
		s.logger.Printf("Processing charge finished: %+v\n", res)
		return res, nil
	}
}

// CreateSession creates a session for a user to pay for a certain product or service.
// This token is intended to allow external interfaces to interact with the payment provider on behalf of the user.
func (s *service) CreateSession(ctx context.Context, req api.CreateSessionRequest) (api.CreateSessionResponse, error) {
	if err := req.Validate(); err != nil {
		return api.CreateSessionResponse{}, err
	}

	return api.CreateSessionResponse{}, nil
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

// Options contains a set of components needed to configure the payment service.
type Options struct {
	// Credits holds a credits.CreditsV1 client implementation.
	Credits credits.CreditsV1

	// Customers holds a customers.CustomersV1 client implementation.
	Customers customers.CustomersV1

	// Stripe holds a stripe API client.
	Stripe *client.API

	// Logger contains a logger mechanism. If set to nil, it defaults to a logger pointing to io.Discard.
	Logger *log.Logger

	// Timeout contains a circuit breaking timeout used to prevent long process runs.
	Timeout time.Duration
}

// NewPaymentsService initializes a new Service implementation using Stripe.
func NewPaymentsService(opts Options) Service {
	if opts.Logger == nil {
		opts.Logger = log.New(io.Discard, "", log.LstdFlags)
	}
	return &service{
		logger:    opts.Logger,
		credits:   opts.Credits,
		customers: opts.Customers,
		timeout:   opts.Timeout,
		stripe:    opts.Stripe,
	}
}
