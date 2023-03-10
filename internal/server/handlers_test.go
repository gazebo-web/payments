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
	"gitlab.com/ignitionrobotics/billing/payments/pkg/adapter"
	"gitlab.com/ignitionrobotics/billing/payments/pkg/api"
	"gitlab.com/ignitionrobotics/billing/payments/pkg/application"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

type handlersTestSuite struct {
	suite.Suite
	Server    *Server
	Credits   *fakecredits.Fake
	Customers *fakecustomers.Fake
	Payments  application.Service
	Logger    *log.Logger
	handler   http.Handler
	Adapter   adapter.Client
	Config    conf.Config
}

func TestStripeWebhookSuite(t *testing.T) {
	suite.Run(t, new(handlersTestSuite))
}

func (s *handlersTestSuite) SetupSuite() {
	s.Logger = log.New(os.Stdout, "[TestStripeWebhook] ", log.LstdFlags|log.Lshortfile|log.Lmsgprefix)
	s.Require().NoError(s.Config.Parse())
}

func (s *handlersTestSuite) SetupTest() {
	s.Credits = fakecredits.NewClient()
	s.Customers = fakecustomers.NewClient()
	s.Adapter = adapter.NewStripeAdapter(s.Config.Stripe)
	s.Payments = application.NewPaymentsService(application.Options{
		Credits:   s.Credits,
		Customers: s.Customers,
		Adapter:   s.Adapter,
		Logger:    s.Logger,
		Timeout:   200 * time.Millisecond,
	})

	var cfg conf.Config
	s.Require().NoError(cfg.Parse())

	s.Server = NewServer(Options{
		config:   cfg,
		payments: s.Payments,
		logger:   s.Logger,
		adapter:  s.Adapter,
	})

}

func (s *handlersTestSuite) TearDownSuite() {
	unsetEnvVars(s.Suite)
}

func (s *handlersTestSuite) TestWebhookEventReceived() {
	s.handler = http.HandlerFunc(s.Server.StripeWebhook)

	body, now := s.prepareEvent(EventPaymentIntentSucceeded, stripe.PaymentIntentStatusSucceeded)

	buff := bytes.NewBuffer(body)

	req, err := http.NewRequest(http.MethodPost, "/", buff)
	s.Require().NoError(err)

	sig := webhook.ComputeSignature(now, body, s.Config.Stripe.SigningKey)
	req.Header.Set("Stripe-Signature", fmt.Sprintf("t=%d,v1=%s", now.Unix(), hex.EncodeToString(sig)))

	rr := httptest.NewRecorder()

	ctx := mock.AnythingOfType("*context.timerCtx")
	user := "test"

	s.Customers.On("GetCustomerByID", ctx, customers.GetCustomerByIDRequest{
		ID:          "cus_CDQTvYK1POcCHA",
		Service:     string(api.PaymentServiceStripe),
		Application: "test",
	}).Return(customers.CustomerResponse{
		Handle:      user,
		ID:          "cus_CDQTvYK1POcCHA",
		Service:     string(api.PaymentServiceStripe),
		Application: "test",
	}, error(nil))

	s.Credits.On("IncreaseCredits", ctx, credits.IncreaseCreditsRequest{
		Transaction: credits.Transaction{
			Handle:      user,
			Application: "test",
			Amount:      100,
			Currency:    "usd",
		},
	}).Return(credits.IncreaseCreditsResponse{}, error(nil))

	s.handler.ServeHTTP(rr, req)

	s.Assert().Equal(http.StatusOK, rr.Code)
}

func (s *handlersTestSuite) TestWebhookGetIdentityFails() {
	s.handler = http.HandlerFunc(s.Server.StripeWebhook)

	body, now := s.prepareEvent(EventPaymentIntentSucceeded, stripe.PaymentIntentStatusSucceeded)

	buff := bytes.NewBuffer(body)

	req, err := http.NewRequest(http.MethodPost, "/", buff)
	s.Require().NoError(err)

	sig := webhook.ComputeSignature(now, body, s.Config.Stripe.SigningKey)
	req.Header.Set("Stripe-Signature", fmt.Sprintf("t=%d,v1=%s", now.Unix(), hex.EncodeToString(sig)))

	rr := httptest.NewRecorder()

	ctx := mock.AnythingOfType("*context.timerCtx")

	s.Customers.On("GetCustomerByID", ctx, customers.GetCustomerByIDRequest{
		ID:          "cus_CDQTvYK1POcCHA",
		Service:     string(api.PaymentServiceStripe),
		Application: "test",
	}).Return(customers.CustomerResponse{}, errors.New("customer service failed"))

	s.handler.ServeHTTP(rr, req)

	s.Assert().Equal(http.StatusInternalServerError, rr.Code)
}

func (s *handlersTestSuite) TestWebhookIncreaseCreditsFails() {
	s.handler = http.HandlerFunc(s.Server.StripeWebhook)

	body, now := s.prepareEvent(EventPaymentIntentSucceeded, stripe.PaymentIntentStatusSucceeded)

	buff := bytes.NewBuffer(body)

	req, err := http.NewRequest(http.MethodPost, "/", buff)
	s.Require().NoError(err)

	sig := webhook.ComputeSignature(now, body, s.Config.Stripe.SigningKey)
	req.Header.Set("Stripe-Signature", fmt.Sprintf("t=%d,v1=%s", now.Unix(), hex.EncodeToString(sig)))

	rr := httptest.NewRecorder()

	ctx := mock.AnythingOfType("*context.timerCtx")
	user := "test"

	s.Customers.On("GetCustomerByID", ctx, customers.GetCustomerByIDRequest{
		ID:          "cus_CDQTvYK1POcCHA",
		Service:     string(api.PaymentServiceStripe),
		Application: "test",
	}).Return(customers.CustomerResponse{
		Handle:      user,
		ID:          "cus_CDQTvYK1POcCHA",
		Service:     string(api.PaymentServiceStripe),
		Application: "test",
	}, error(nil))

	s.Credits.On("IncreaseCredits", ctx, credits.IncreaseCreditsRequest{
		Transaction: credits.Transaction{
			Handle:      user,
			Application: "test",
			Amount:      100,
			Currency:    "usd",
		},
	}).Return(credits.IncreaseCreditsResponse{}, errors.New("credits service failed"))

	s.handler.ServeHTTP(rr, req)

	s.Assert().Equal(http.StatusInternalServerError, rr.Code)
}

func (s *handlersTestSuite) TestWebhookTimeout() {
	s.handler = http.HandlerFunc(s.Server.StripeWebhook)

	body, now := s.prepareEvent(EventPaymentIntentSucceeded, stripe.PaymentIntentStatusSucceeded)

	buff := bytes.NewBuffer(body)

	req, err := http.NewRequest(http.MethodPost, "/", buff)
	s.Require().NoError(err)

	sig := webhook.ComputeSignature(now, body, s.Config.Stripe.SigningKey)
	req.Header.Set("Stripe-Signature", fmt.Sprintf("t=%d,v1=%s", now.Unix(), hex.EncodeToString(sig)))

	rr := httptest.NewRecorder()

	ctx := mock.AnythingOfType("*context.timerCtx")
	user := "test"

	s.Customers.On("GetCustomerByID", ctx, customers.GetCustomerByIDRequest{
		ID:          "cus_CDQTvYK1POcCHA",
		Service:     string(api.PaymentServiceStripe),
		Application: "test",
	}).Return(customers.CustomerResponse{
		Handle:      user,
		ID:          "cus_CDQTvYK1POcCHA",
		Service:     string(api.PaymentServiceStripe),
		Application: "test",
	}, error(nil))

	s.Credits.On("IncreaseCredits", ctx, credits.IncreaseCreditsRequest{
		Transaction: credits.Transaction{
			Handle:      user,
			Application: "test",
			Amount:      100,
			Currency:    "usd",
		},
	}).Return(credits.IncreaseCreditsResponse{}, error(nil)).Run(func(args mock.Arguments) {
		time.Sleep(1 * time.Second)
	})

	s.handler.ServeHTTP(rr, req)

	s.Assert().Equal(http.StatusInternalServerError, rr.Code)
}

func (s *handlersTestSuite) TestWebhookEventFailed() {
	s.handler = http.HandlerFunc(s.Server.StripeWebhook)

	body, now := s.prepareEvent(EventPaymentIntentFailed, stripe.PaymentIntentStatusCanceled)

	buff := bytes.NewBuffer(body)

	req, err := http.NewRequest(http.MethodPost, "/", buff)
	s.Require().NoError(err)

	sig := webhook.ComputeSignature(now, body, s.Config.Stripe.SigningKey)
	req.Header.Set("Stripe-Signature", fmt.Sprintf("t=%d,v1=%s", now.Unix(), hex.EncodeToString(sig)))

	rr := httptest.NewRecorder()

	s.handler.ServeHTTP(rr, req)

	s.Assert().Equal(http.StatusInternalServerError, rr.Code)
}

func (s *handlersTestSuite) TestCreateSessionOK() {
	s.handler = http.HandlerFunc(s.Server.CreateSession)

	body, err := json.Marshal(api.CreateSessionRequest{
		Service:     "stripe",
		SuccessURL:  "http://localhost",
		CancelURL:   "http://localhost",
		Handle:      "test",
		Application: "test",
	})

	buff := bytes.NewBuffer(body)
	req, err := http.NewRequest(http.MethodPost, "/", buff)
	s.Require().NoError(err)

	rr := httptest.NewRecorder()

	ctx := mock.AnythingOfType("*context.timerCtx")
	s.Customers.On("GetCustomerByHandle", ctx, customers.GetCustomerByHandleRequest{
		Handle:      "test",
		Service:     string(api.PaymentServiceStripe),
		Application: "test",
	}).Return(customers.CustomerResponse{
		Handle:      "test",
		Service:     string(api.PaymentServiceStripe),
		Application: "test",
		ID:          "cus_HdRJTeoStCxpP4E",
	}, error(nil))

	s.Credits.On("GetUnitPrice", ctx, credits.GetUnitPriceRequest{Currency: "usd"}).Return(credits.GetUnitPriceResponse{
		Amount:   2,
		Currency: "usd",
	}, error(nil))

	s.handler.ServeHTTP(rr, req)

	body, err = io.ReadAll(rr.Body)
	s.Require().NoError(err)

	var out api.CreateSessionResponse
	s.Require().NoError(json.Unmarshal(body, &out))

	s.Assert().Equal(http.StatusOK, rr.Code)
	s.Assert().NotEmpty(out.Session)
	s.Assert().Equal(api.PaymentServiceStripe, out.Service)
}

func (s *handlersTestSuite) prepareEvent(eventType string, status stripe.PaymentIntentStatus) ([]byte, time.Time) {
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
