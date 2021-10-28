package server

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/webhook"
	credits "gitlab.com/ignitionrobotics/billing/credits/pkg/api"
	fakecredits "gitlab.com/ignitionrobotics/billing/credits/pkg/fake"
	customers "gitlab.com/ignitionrobotics/billing/customers/pkg/api"
	fakecustomers "gitlab.com/ignitionrobotics/billing/customers/pkg/fake"
	"gitlab.com/ignitionrobotics/billing/payments/internal/conf"
	"gitlab.com/ignitionrobotics/billing/payments/pkg/api"
	"gitlab.com/ignitionrobotics/billing/payments/pkg/application"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

type stripeWebhookSuite struct {
	suite.Suite
	Server    *Server
	Credits   *fakecredits.Fake
	Customers *fakecustomers.Fake
	Payments  application.Service
	Logger    *log.Logger
	handler   http.Handler
}

func TestStripeWebhookSuite(t *testing.T) {
	suite.Run(t, new(stripeWebhookSuite))
}

func (s *stripeWebhookSuite) SetupTest() {
	s.Logger = log.New(os.Stdout, "[TestStripeWebhook] ", log.LstdFlags|log.Lshortfile|log.Lmsgprefix)

	s.Credits = fakecredits.NewClient()
	s.Customers = fakecustomers.NewClient()
	s.Payments = application.NewPaymentsService(s.Credits, s.Customers, s.Logger, 200*time.Millisecond)
	s.Server = NewServer(Options{
		config: conf.Config{
			Port:             80,
			StripeSigningKey: "whsec_test1234",
		},
		payments: s.Payments,
		logger:   s.Logger,
	})
	s.handler = http.HandlerFunc(s.Server.StripeWebhook)
}

func (s *stripeWebhookSuite) TestWebhookEventReceived() {
	body, now := s.prepareEvent(EventPaymentIntentSucceeded, stripe.PaymentIntentStatusSucceeded)

	buff := bytes.NewBuffer(body)

	req, err := http.NewRequest(http.MethodPost, "/", buff)
	s.Require().NoError(err)

	sig := webhook.ComputeSignature(now, body, "whsec_test1234")
	req.Header.Set("Stripe-Signature", fmt.Sprintf("t=%d,v1=%s", now.Unix(), hex.EncodeToString(sig)))

	rr := httptest.NewRecorder()

	ctx := mock.AnythingOfType("*context.timerCtx")
	user := "test"

	s.Customers.On("GetIdentity", ctx, customers.GetIdentityRequest{
		Customer:    "cus_CDQTvYK1POcCHA",
		Service:     string(api.PaymentServiceStripe),
		Application: "test",
	}).Return(customers.GetIdentityResponse{
		User:        &user,
		Customer:    nil,
		Service:     string(api.PaymentServiceStripe),
		Application: "test",
	}, error(nil))

	s.Credits.On("IncreaseCredits", ctx, credits.IncreaseCreditsRequest{
		User:     user,
		Amount:   100,
		Currency: "usd",
	}).Return(credits.IncreaseCreditsResponse{}, error(nil))

	s.handler.ServeHTTP(rr, req)

	s.Assert().Equal(http.StatusOK, rr.Code)
}

func (s *stripeWebhookSuite) TestWebhookGetIdentityFails() {
	body, now := s.prepareEvent(EventPaymentIntentSucceeded, stripe.PaymentIntentStatusSucceeded)

	buff := bytes.NewBuffer(body)

	req, err := http.NewRequest(http.MethodPost, "/", buff)
	s.Require().NoError(err)

	sig := webhook.ComputeSignature(now, body, "whsec_test1234")
	req.Header.Set("Stripe-Signature", fmt.Sprintf("t=%d,v1=%s", now.Unix(), hex.EncodeToString(sig)))

	rr := httptest.NewRecorder()

	ctx := mock.AnythingOfType("*context.timerCtx")

	s.Customers.On("GetIdentity", ctx, customers.GetIdentityRequest{
		Customer:    "cus_CDQTvYK1POcCHA",
		Service:     string(api.PaymentServiceStripe),
		Application: "test",
	}).Return(customers.GetIdentityResponse{}, errors.New("identity service failed"))

	s.handler.ServeHTTP(rr, req)

	s.Assert().Equal(http.StatusInternalServerError, rr.Code)
}

func (s *stripeWebhookSuite) TestWebhookIncreaseCreditsFails() {
	body, now := s.prepareEvent(EventPaymentIntentSucceeded, stripe.PaymentIntentStatusSucceeded)

	buff := bytes.NewBuffer(body)

	req, err := http.NewRequest(http.MethodPost, "/", buff)
	s.Require().NoError(err)

	sig := webhook.ComputeSignature(now, body, "whsec_test1234")
	req.Header.Set("Stripe-Signature", fmt.Sprintf("t=%d,v1=%s", now.Unix(), hex.EncodeToString(sig)))

	rr := httptest.NewRecorder()

	ctx := mock.AnythingOfType("*context.timerCtx")
	user := "test"

	s.Customers.On("GetIdentity", ctx, customers.GetIdentityRequest{
		Customer:    "cus_CDQTvYK1POcCHA",
		Service:     string(api.PaymentServiceStripe),
		Application: "test",
	}).Return(customers.GetIdentityResponse{
		User:        &user,
		Customer:    nil,
		Service:     string(api.PaymentServiceStripe),
		Application: "test",
	}, error(nil))

	s.Credits.On("IncreaseCredits", ctx, credits.IncreaseCreditsRequest{
		User:     user,
		Amount:   100,
		Currency: "usd",
	}).Return(credits.IncreaseCreditsResponse{}, errors.New("credits service failed"))

	s.handler.ServeHTTP(rr, req)

	s.Assert().Equal(http.StatusInternalServerError, rr.Code)
}

func (s *stripeWebhookSuite) TestWebhookTimeout() {
	body, now := s.prepareEvent(EventPaymentIntentSucceeded, stripe.PaymentIntentStatusSucceeded)

	buff := bytes.NewBuffer(body)

	req, err := http.NewRequest(http.MethodPost, "/", buff)
	s.Require().NoError(err)

	sig := webhook.ComputeSignature(now, body, "whsec_test1234")
	req.Header.Set("Stripe-Signature", fmt.Sprintf("t=%d,v1=%s", now.Unix(), hex.EncodeToString(sig)))

	rr := httptest.NewRecorder()

	ctx := mock.AnythingOfType("*context.timerCtx")
	user := "test"

	s.Customers.On("GetIdentity", ctx, customers.GetIdentityRequest{
		Customer:    "cus_CDQTvYK1POcCHA",
		Service:     string(api.PaymentServiceStripe),
		Application: "test",
	}).Return(customers.GetIdentityResponse{
		User:        &user,
		Customer:    nil,
		Service:     string(api.PaymentServiceStripe),
		Application: "test",
	}, error(nil))

	s.Credits.On("IncreaseCredits", ctx, credits.IncreaseCreditsRequest{
		User:     user,
		Amount:   100,
		Currency: "usd",
	}).Return(credits.IncreaseCreditsResponse{}, error(nil)).Run(func(args mock.Arguments) {
		time.Sleep(1 * time.Second)
	})

	s.handler.ServeHTTP(rr, req)

	s.Assert().Equal(http.StatusInternalServerError, rr.Code)
}

func (s *stripeWebhookSuite) TestWebhookEventFailed() {
	body, now := s.prepareEvent(EventPaymentIntentFailed, stripe.PaymentIntentStatusCanceled)

	buff := bytes.NewBuffer(body)

	req, err := http.NewRequest(http.MethodPost, "/", buff)
	s.Require().NoError(err)

	sig := webhook.ComputeSignature(now, body, "whsec_test1234")
	req.Header.Set("Stripe-Signature", fmt.Sprintf("t=%d,v1=%s", now.Unix(), hex.EncodeToString(sig)))

	rr := httptest.NewRecorder()

	s.handler.ServeHTTP(rr, req)

	s.Assert().Equal(http.StatusBadRequest, rr.Code)
}

func (s *stripeWebhookSuite) prepareEvent(eventType string, status stripe.PaymentIntentStatus) ([]byte, time.Time) {
	now := time.Now()

	data, err := json.Marshal(stripe.PaymentIntent{
		Amount:   100,
		Created:  now.Unix(),
		Currency: "usd",
		Customer: &stripe.Customer{
			ID:    "cus_CDQTvYK1POcCHA",
			Email: "robot@test.org",
		},
		Description:  "A test description for a payment intent",
		ID:           "pi_5DpcTV1eZvKYlo3Cy7cIe9am",
		ReceiptEmail: "robot@test.org",
		Status:       status,
		Metadata: map[string]string{
			"application": "test",
		},
	})
	s.Require().NoError(err)

	eventData := stripe.EventData{
		Raw: data,
	}

	event := stripe.Event{
		Created: now.Unix(),
		Data:    &eventData,
		ID:      "evt_1CiPtv2eZvKYlo2CcUZsDcO6",
		Type:    eventType,
	}

	body, err := json.Marshal(event)
	s.Require().NoError(err)

	return body, now
}
