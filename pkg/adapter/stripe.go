package adapter

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/client"
	"github.com/stripe/stripe-go/v72/webhook"
	customers "gitlab.com/ignitionrobotics/billing/customers/pkg/api"
	"gitlab.com/ignitionrobotics/billing/payments/internal/conf"
	"gitlab.com/ignitionrobotics/billing/payments/pkg/api"
)

const (
	// EventPaymentIntentSucceeded is the event triggered by Stripe when a payment intent succeeds.
	EventPaymentIntentSucceeded = "payment_intent.succeeded"
)

// stripeAdapter implements Client using the Stripe API and tools.
type stripeAdapter struct {
	// SigningKey is used to validate a webhook event.
	SigningKey string
	// API contains a stripe client implementation.
	API *client.API
}

// GenerateChargeRequest generates an api.ChargeRequest out from the given body a set of parameters.
func (s *stripeAdapter) GenerateChargeRequest(body []byte, params map[string][]string) (api.ChargeRequest, error) {
	// Get stripe signature
	var sig string
	if p, ok := params["Stripe-Signature"]; !ok || len(p) == 0 {
		return api.ChargeRequest{}, errors.New("invalid signature")
	} else {
		sig = p[0]
	}

	// Validate event
	event, err := webhook.ConstructEvent(body, sig, s.SigningKey)
	if err != nil {
		return api.ChargeRequest{}, err
	}

	// Check event is a payment intent succeeded
	if event.Type != EventPaymentIntentSucceeded {
		return api.ChargeRequest{}, errors.New("couldn't process event type")
	}

	// Parse payment intent
	var paymentIntent stripe.PaymentIntent
	if err = json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
		return api.ChargeRequest{}, err
	}

	// A customer should be defined
	if paymentIntent.Customer == nil {
		return api.ChargeRequest{}, errors.New("missing customer")
	}

	// Get application metadata
	var app string
	var ok bool
	if app, ok = paymentIntent.Metadata["application"]; !ok {
		return api.ChargeRequest{}, errors.New("missing application")
	}

	// Parse charge
	return api.ChargeRequest{
		Amount:      uint(paymentIntent.Amount),
		Currency:    paymentIntent.Currency,
		Customer:    paymentIntent.Customer.ID,
		Service:     api.PaymentServiceStripe,
		Application: app,
	}, nil
}

// CreateCustomer creates a customer in Stripe for the given application. It returns the ID of the new customer.
// Stripe docs: https://stripe.com/docs/api/customers/create
func (s *stripeAdapter) CreateCustomer(application, handle string) (string, error) {
	c, err := s.API.Customers.New(&stripe.CustomerParams{
		Description: stripe.String(fmt.Sprintf("Customer (%s) created for application: %s", handle, application)),
	})
	if err != nil {
		return "", nil
	}
	return c.ID, nil
}

// CreateSession initializes a new Stripe Checkout session.
// Stripe docs: https://stripe.com/docs/api/checkout/sessions/create
func (s *stripeAdapter) CreateSession(req api.CreateSessionRequest, cus customers.CustomerResponse) (api.CreateSessionResponse, error) {
	session, err := s.API.CheckoutSessions.New(&stripe.CheckoutSessionParams{
		SuccessURL: &req.SuccessURL,
		CancelURL:  &req.CancelURL,
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		Customer: stripe.String(cus.ID),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Quantity: stripe.Int64(1),
				AdjustableQuantity: &stripe.CheckoutSessionLineItemAdjustableQuantityParams{
					Enabled: stripe.Bool(true),
					Maximum: stripe.Int64(1000), // Max amount of credits to buy
					Minimum: stripe.Int64(1),    // Min amount of credits to buy
				},
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("usd"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("Credits"),
					},
					UnitAmount: stripe.Int64(100), // Price per credit
				},
			},
		},
		Params: stripe.Params{
			Metadata: map[string]string{
				"application": req.Application, // Used by webhooks
				"handle":      req.Handle,
			},
		},
	})
	if err != nil {
		return api.CreateSessionResponse{}, err
	}
	return api.CreateSessionResponse{
		Service: req.Service,
		Session: session.ID,
	}, nil
}

// NewStripeAdapter initializes a new adapter using the Stripe client.
func NewStripeAdapter(cfg conf.Stripe) Client {
	var backendURL *string
	if len(cfg.URL) > 0 {
		backendURL = &cfg.URL
	}
	config := stripe.BackendConfig{
		URL: backendURL,
	}
	c := client.New(cfg.SecretKey, &stripe.Backends{
		API:     stripe.GetBackendWithConfig(stripe.APIBackend, &config),
		Connect: stripe.GetBackendWithConfig(stripe.ConnectBackend, &config),
		Uploads: stripe.GetBackendWithConfig(stripe.UploadsBackend, &config),
	})
	return &stripeAdapter{
		SigningKey: cfg.SigningKey,
		API:        c,
	}
}
