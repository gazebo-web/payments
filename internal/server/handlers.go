package server

import (
	"encoding/json"
	"fmt"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/webhook"
	"gitlab.com/ignitionrobotics/billing/payments/pkg/api"
	"io"
	"net/http"
)

const (
	// EventPaymentIntentSucceeded is the event triggered by Stripe when a payment intent succeeds.
	EventPaymentIntentSucceeded = "payment_intent.succeeded"
	// EventPaymentIntentFailed is the event triggered by Stripe when a payment intent failed.
	EventPaymentIntentFailed = "payment_intent.payment_failed"
)

// StripeWebhook is used for receiving Stripe webhook events related to payment intents and process those events to add credits
// to specific users.
// 	Each payment intent event should contain a valid customer and should include a metadata value with the application name.
// 	Example:
//		"application": "fuel"
func (s *Server) StripeWebhook(w http.ResponseWriter, r *http.Request) {
	s.logger.Println("Reading request body")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s - %s", http.StatusText(http.StatusInternalServerError), "Failed to read body"), http.StatusInternalServerError)
		return
	}

	s.logger.Println("Parsing incoming Stripe webhook event")
	event, err := webhook.ConstructEvent(body, r.Header.Get("Stripe-Signature"), s.webhookStripeSigningKey)
	if err != nil {
		s.logger.Println("Signed event is invalid:", err)
		http.Error(w, fmt.Sprintf("%s - %s", http.StatusText(http.StatusForbidden), "Invalid event"), http.StatusForbidden)
		return
	}

	s.logger.Println("Checking payment status")
	if event.Type != EventPaymentIntentSucceeded {
		s.logger.Println("The payment did not succeed:", event.Type)
		http.Error(w, fmt.Sprintf("%s - %s", http.StatusText(http.StatusBadRequest), "Couldn't process event type"), http.StatusBadRequest)
		return
	}

	s.logger.Println("Unmarshalling payment intent")
	var paymentIntent stripe.PaymentIntent
	if err = json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
		s.logger.Println("Failed to unmarshal payment intent:", err)
		http.Error(w, fmt.Sprintf("%s - %s", http.StatusText(http.StatusBadRequest), "Couldn't process event data"), http.StatusBadRequest)
		return
	}

	s.logger.Println("Checking customer is defined")
	if paymentIntent.Customer == nil {
		s.logger.Println("Failed to get customer")
		http.Error(w, fmt.Sprintf("%s - %s", http.StatusText(http.StatusBadRequest), "Missing customer"), http.StatusBadRequest)
		return
	}

	s.logger.Println("Getting application id from metadata")
	var app string
	var ok bool
	if app, ok = paymentIntent.Metadata["application"]; !ok {
		s.logger.Println("Failed to get application ID")
		http.Error(w, fmt.Sprintf("%s - %s", http.StatusText(http.StatusBadRequest), "Missing application"), http.StatusBadRequest)
		return
	}

	_, err = s.payments.Charge(r.Context(), api.ChargeRequest{
		Amount:      uint(paymentIntent.Amount),
		Currency:    paymentIntent.Currency,
		Customer:    paymentIntent.Customer.ID,
		Service:     api.PaymentServiceStripe,
		Application: app,
	})
	if err != nil {
		s.logger.Println("Failed to process charge:", err)
		http.Error(w, fmt.Sprintf("%s - %s: %v", http.StatusText(http.StatusInternalServerError), "Failed to charge customer", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write([]byte(fmt.Sprintf("%s - Payment processed", http.StatusText(http.StatusCreated)))); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
