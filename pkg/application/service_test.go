package application

import (
	"context"
	"github.com/stretchr/testify/suite"
	fakecredits "gitlab.com/ignitionrobotics/billing/credits/pkg/fake"
	fakecustomers "gitlab.com/ignitionrobotics/billing/customers/pkg/fake"
	"gitlab.com/ignitionrobotics/billing/payments/pkg/api"
	"testing"
	"time"
)

type serviceTestSuite struct {
	suite.Suite
	Credits   *fakecredits.Fake
	Customers *fakecustomers.Fake
	Service   Service
}

func TestPaymentsService(t *testing.T) {
	suite.Run(t, new(serviceTestSuite))
}

func (s *serviceTestSuite) SetupTest() {
	s.Credits = fakecredits.NewClient()
	s.Customers = fakecustomers.NewClient()
	s.Service = NewPaymentsService(s.Credits, s.Customers, nil, 200*time.Millisecond)
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
