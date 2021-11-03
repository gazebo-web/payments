package application

import (
	"context"
	"github.com/stretchr/testify/suite"
	"github.com/stripe/stripe-go/v72/client"
	fakecredits "gitlab.com/ignitionrobotics/billing/credits/pkg/fake"
	fakecustomers "gitlab.com/ignitionrobotics/billing/customers/pkg/fake"
	"gitlab.com/ignitionrobotics/billing/payments/internal/conf"
	"gitlab.com/ignitionrobotics/billing/payments/pkg/api"
	"testing"
	"time"
)

type serviceTestSuite struct {
	suite.Suite
	Credits   *fakecredits.Fake
	Customers *fakecustomers.Fake
	Service   Service
	Stripe    *client.API
	StripeURL string
}

func TestPaymentsService(t *testing.T) {
	suite.Run(t, new(serviceTestSuite))
}

func (s *serviceTestSuite) SetupTest() {
	s.Credits = fakecredits.NewClient()
	s.Customers = fakecustomers.NewClient()
	s.StripeURL = "http://stripe:12111"

	var cfg conf.Config
	s.Require().NoError(cfg.Parse())

	s.Stripe = NewStripeClient(cfg.Stripe)
	s.Service = NewPaymentsService(Options{
		Credits:   s.Credits,
		Customers: s.Customers,
		Stripe:    s.Stripe,
		Timeout:   200 * time.Millisecond,
	})
}

func (s *serviceTestSuite) TestCreateSessionServiceIsEmpty() {
	_, err := s.Service.CreateSession(context.Background(), api.CreateSessionRequest{
		Service: "",
	})
	s.Assert().Error(err)
	s.Assert().Equal(api.ErrEmptyService, err)
}

func (s *serviceTestSuite) TestCreateSessionServiceIsNotValid() {
	_, err := s.Service.CreateSession(context.Background(), api.CreateSessionRequest{
		Service: "paypal",
	})
	s.Assert().Error(err)
	s.Assert().Equal(api.ErrInvalidService, err)
}

func (s *serviceTestSuite) TestCreateSessionEmptyURLs() {
	_, err := s.Service.CreateSession(context.Background(), api.CreateSessionRequest{
		Service:    api.PaymentServiceStripe,
		SuccessURL: "https://localhost/success",
		CancelURL:  "",
	})
	s.Assert().Error(err)
	s.Assert().Equal(api.ErrEmptyCallbacks, err)

	_, err = s.Service.CreateSession(context.Background(), api.CreateSessionRequest{
		Service:    api.PaymentServiceStripe,
		SuccessURL: "",
		CancelURL:  "https://localhost/cancel",
	})
	s.Assert().Error(err)
	s.Assert().Equal(api.ErrEmptyCallbacks, err)

	_, err = s.Service.CreateSession(context.Background(), api.CreateSessionRequest{
		Service:    api.PaymentServiceStripe,
		SuccessURL: "",
		CancelURL:  "",
	})
	s.Assert().Error(err)
	s.Assert().Equal(api.ErrEmptyCallbacks, err)
}

func (s *serviceTestSuite) TestCreateSessionURLsMalformed() {
	_, err := s.Service.CreateSession(context.Background(), api.CreateSessionRequest{
		Service:    api.PaymentServiceStripe,
		SuccessURL: "https://localhost",
		CancelURL:  "testinginvalidurl",
	})
	s.Assert().Error(err)

	_, err = s.Service.CreateSession(context.Background(), api.CreateSessionRequest{
		Service:    api.PaymentServiceStripe,
		SuccessURL: "testinginvalidurl",
		CancelURL:  "https://localhost",
	})
	s.Assert().Error(err)
}
