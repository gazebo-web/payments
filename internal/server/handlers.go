package server

import (
	"fmt"
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
		s.logger.Println("Failed to read body:", err)
		http.Error(w, fmt.Sprintf("%s - %s", http.StatusText(http.StatusInternalServerError), "Failed to read body"), http.StatusInternalServerError)
		return
	}

	req, err := s.adapter.GenerateChargeRequest(body, r.Header)
	if err != nil {
		s.logger.Println("Failed to generate charge request:", err)
		http.Error(w, fmt.Sprintf("%s - %s: %v", http.StatusText(http.StatusInternalServerError), "Failed to generate charge request", err), http.StatusInternalServerError)
		return
	}

	_, err = s.payments.Charge(r.Context(), req)
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

// CreateSession is an HTTP handler to call the api.PaymentsV1's CreateSession method.
func (s *Server) CreateSession(w http.ResponseWriter, r *http.Request) {

}
