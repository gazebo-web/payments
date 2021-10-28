package api

import (
	"context"
)

// PaymentService identifies different payment services such as Stripe, PayPal, and more.
type PaymentService string

const (
	// PaymentServiceStripe represents the stripe payment service.
	PaymentServiceStripe PaymentService = "stripe"
)

// ChargerV1 contains methods that should be called after charging a certain amount of money to a user.
// This interface is private to the payment service and should only be called from a webhook 
// after a payment system event is processed.
type ChargerV1 interface {
	// Charge charges a certain amount of money to a given user.
	Charge(ctx context.Context, req ChargeRequest) (ChargeResponse, error)
}

// ChargeRequest is the input for the ChargerV1.Charge method.
type ChargeRequest struct {
	// Amount contains the value in Cents that has been charged to a certain user.
	Amount uint

	// Currency holds the ISO 4217 currency value in lowercase format.
	//	Examples: usd, eur.
	Currency string

	// Customer contains a value that represents a user in a certain payment system.
	Customer string

	// Service contains the name of the payment service that has been used to perform this charge.
	Service PaymentService

	// Application contains an identifier of an application that originated this charge.
	Application string
}

// ChargeResponse is the output of the ChargerV1.Charge method.
type ChargeResponse struct{}

// PaymentsV1 holds the methods that allow interacting with a payment platform such as Stripe.
// The audience of this interface is internal to the different billing and application services as it shouldn't be called
// from the internet.
type PaymentsV1 interface {
	// CreateSession creates a session for a user to pay for a certain product or service.
	CreateSession(ctx context.Context, req CreateSessionRequest) (CreateSessionResponse, error)

	// ListInvoices returns a list of invoices of the given user.
	ListInvoices(ctx context.Context, req ListInvoicesRequest) (ListInvoicesResponse, error)
}

// CreateSessionRequest is the input for the PaymentsV1.CreateSession method.
type CreateSessionRequest struct{}

// CreateSessionResponse is the output of the PaymentsV1.CreateSession method.
type CreateSessionResponse struct{}

// ListInvoicesRequest is the input for the PaymentsV1.ListInvoices method.
type ListInvoicesRequest struct{}

// ListInvoicesResponse is the output of the PaymentsV1.ListInvoices method.
type ListInvoicesResponse struct{}
