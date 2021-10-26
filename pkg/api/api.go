package api

import (
	"context"
)

type StatusCharge string

const (
	StatusChargeSucceeded     StatusCharge = "Succeeded"
	StatusChargePaymentFailed StatusCharge = "PaymentFailed"
)

// ChargerV1 holds the methods to be called after charging a certain amount of money to user.
type ChargerV1 interface {
	// Charge charges a certain amount of money to a given user.
	Charge(ctx context.Context, req ChargeRequest) (ChargeResponse, error)
}

// ChargeRequest is the input for the ChargerV1.Charge method.
type ChargeRequest struct{}

// ChargeResponse is the output of the ChargerV1.Charge method.
type ChargeResponse struct{}

// PaymentsV1 holds the methods that allow interacting with a payment platform such as Stripe.
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
