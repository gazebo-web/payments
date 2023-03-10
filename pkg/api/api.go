package api

import (
	"context"
	"errors"
)

var (
	// ErrEmptyService is returned when the service provided is empty.
	ErrEmptyService = errors.New("empty service")

	// ErrInvalidService is returned when an invalid service is used.
	ErrInvalidService = errors.New("invalid service")

	// ErrEmptyCallbacks is returned when either (or both) of the callback URLs are empty.
	ErrEmptyCallbacks = errors.New("empty callbacks")

	// ErrInvalidURL is returned when an invalid URL callback value is passed on a request.
	ErrInvalidURL = errors.New("invalid URL")

	// ErrEmptyHandle is returned when an empty handle value is passed on a request.
	ErrEmptyHandle = errors.New("empty handle")

	// ErrEmptyApplication is returned when an empty application value is passed on a request.
	ErrEmptyApplication = errors.New("empty application")

	// ErrInvalidUnitPrice is returned when an invalid unit price is passed on a request.
	ErrInvalidUnitPrice = errors.New("invalid unit price")
)

// PaymentService identifies different payment services such as Stripe, PayPal, and more.
type PaymentService string

const (
	// PaymentServiceStripe represents the stripe payment service.
	PaymentServiceStripe PaymentService = "stripe"
)

// Validate validates the current payment service.
func (ps PaymentService) Validate() error {
	if len(ps) == 0 {
		return ErrEmptyService
	}
	if ps != PaymentServiceStripe {
		return ErrInvalidService
	}
	return nil
}

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
type CreateSessionRequest struct {
	// Service contains the name of the payment service that should be used to start a transaction session.
	Service PaymentService `json:"service"`

	// SuccessURL is the URL where to redirect a checkout process when it succeeds.
	SuccessURL string `json:"success_url"`

	// CancelURL is the URL where to redirect a checkout process when it fails.
	CancelURL string `json:"cancel_url"`

	// Handle is the customer identity in the context of a certain application.
	// E.g. application username, application organization name.
	Handle string `json:"handle"`

	// Application is the application that requested the creation of this session.
	Application string `json:"application"`

	// UnitPrice is the amount of cents a credit costs.
	// This field is ignored.
	// TODO: Remove this field from the public-facing API data structure.
	UnitPrice uint `json:"-"`
}

// Validate validates the current request.
func (r CreateSessionRequest) Validate() error {
	if err := r.Service.Validate(); err != nil {
		return err
	}
	if len(r.SuccessURL) == 0 || len(r.CancelURL) == 0 {
		return ErrEmptyCallbacks
	}

	if err := validateURL(r.SuccessURL); err != nil {
		return err
	}

	if err := validateURL(r.CancelURL); err != nil {
		return err
	}

	if len(r.Handle) == 0 {
		return ErrEmptyHandle
	}

	if len(r.Application) == 0 {
		return ErrEmptyApplication
	}

	if r.UnitPrice == 0 {
		return ErrInvalidUnitPrice
	}

	return nil
}

// CreateSessionResponse is the output of the PaymentsV1.CreateSession method.
type CreateSessionResponse struct {
	// Service contains the name of the service where the transaction is taking place.
	Service PaymentService `json:"service"`

	// Session is the ID of the session created for this transaction.
	Session string `json:"session"`
}

// ListInvoicesRequest is the input for the PaymentsV1.ListInvoices method.
type ListInvoicesRequest struct{}

// ListInvoicesResponse is the output of the PaymentsV1.ListInvoices method.
type ListInvoicesResponse struct{}
