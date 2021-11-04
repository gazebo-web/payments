package application

import (
	"context"
	"errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/stripe/stripe-go/v72/client"
	fakecredits "gitlab.com/ignitionrobotics/billing/credits/pkg/fake"
	customers "gitlab.com/ignitionrobotics/billing/customers/pkg/api"
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
}

func TestPaymentsService(t *testing.T) {
	suite.Run(t, new(serviceTestSuite))
}

func (s *serviceTestSuite) SetupTest() {
	s.Credits = fakecredits.NewClient()
	s.Customers = fakecustomers.NewClient()

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

func (s *serviceTestSuite) TestCreateSessionEmptyHandle() {
	_, err := s.Service.CreateSession(context.Background(), api.CreateSessionRequest{
		Service:    api.PaymentServiceStripe,
		SuccessURL: "https://localhost",
		CancelURL:  "https://localhost",
		Handle:     "",
	})
	s.Assert().Error(err)
	s.Assert().Equal(err, api.ErrEmptyHandle)
}

func (s *serviceTestSuite) TestCreateSessionEmptyApplication() {
	_, err := s.Service.CreateSession(context.Background(), api.CreateSessionRequest{
		Service:     api.PaymentServiceStripe,
		SuccessURL:  "https://localhost",
		CancelURL:   "https://localhost",
		Handle:      "test",
		Application: "",
	})
	s.Assert().Error(err)
	s.Assert().Equal(err, api.ErrEmptyApplication)
}

func (s *serviceTestSuite) TestCreateSessionCustomerFailed() {
	ctx := mock.AnythingOfType("*context.timerCtx")
	s.Customers.On("GetCustomerByHandle", ctx, customers.GetCustomerByHandleRequest{
		Handle:      "test",
		Service:     string(api.PaymentServiceStripe),
		Application: "test",
	}).Return(customers.CustomerResponse{}, errors.New("customer service failed"))

	_, err := s.Service.CreateSession(context.Background(), api.CreateSessionRequest{
		Service:     api.PaymentServiceStripe,
		SuccessURL:  "https://localhost",
		CancelURL:   "https://localhost",
		Handle:      "test",
		Application: "test",
	})
	s.Assert().Error(err)
}

func (s *serviceTestSuite) TestCreateSessionOKWithCustomerCreation() {
	ctx := mock.AnythingOfType("*context.timerCtx")
	s.Customers.On("GetCustomerByHandle", ctx, customers.GetCustomerByHandleRequest{
		Handle:      "test",
		Service:     string(api.PaymentServiceStripe),
		Application: "test",
	}).Return(customers.CustomerResponse{}, customers.ErrCustomerNotFound)

	s.Customers.On("CreateCustomer", ctx, mock.AnythingOfType("api.CreateCustomerRequest")).Return(customers.CustomerResponse{
		Handle:      "test",
		Service:     string(api.PaymentServiceStripe),
		Application: "test",
		ID:          "cus_HdRJTeoStCxpP4E",
	}, error(nil))

	res, err := s.Service.CreateSession(context.Background(), api.CreateSessionRequest{
		Service:     api.PaymentServiceStripe,
		SuccessURL:  "https://localhost",
		CancelURL:   "https://localhost",
		Handle:      "test",
		Application: "test",
	})
	s.Require().NoError(err)

	s.Assert().Equal(api.PaymentServiceStripe, res.Service)
	s.Assert().NotEmpty(res.Session)
}

func (s *serviceTestSuite) TestCreateSessionOK() {
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

	res, err := s.Service.CreateSession(context.Background(), api.CreateSessionRequest{
		Service:     api.PaymentServiceStripe,
		SuccessURL:  "https://localhost",
		CancelURL:   "https://localhost",
		Handle:      "test",
		Application: "test",
	})
	s.Require().NoError(err)

	s.Assert().Equal(api.PaymentServiceStripe, res.Service)
	s.Assert().NotEmpty(res.Session)
}
