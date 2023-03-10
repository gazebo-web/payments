package server

import (
	"context"
	"github.com/stretchr/testify/suite"
	credits "gitlab.com/ignitionrobotics/billing/credits/pkg/client"
	fakecredits "gitlab.com/ignitionrobotics/billing/credits/pkg/fake"
	customers "gitlab.com/ignitionrobotics/billing/customers/pkg/client"
	fakecustomers "gitlab.com/ignitionrobotics/billing/customers/pkg/fake"
	"gitlab.com/ignitionrobotics/billing/payments/internal/conf"
	"gitlab.com/ignitionrobotics/billing/payments/pkg/adapter"
	"gitlab.com/ignitionrobotics/billing/payments/pkg/application"
	"log"
	"os"
	"testing"
	"time"
)

type setupTestSuite struct {
	suite.Suite
	Logger *log.Logger
}

func TestSetupSuite(t *testing.T) {
	suite.Run(t, new(setupTestSuite))
}

func (s *setupTestSuite) SetupSuite() {
	s.Logger = log.New(os.Stdout, "[TestSetup] ", log.LstdFlags|log.Lshortfile|log.Lmsgprefix)
}

func (s *setupTestSuite) TearDownTest() {
	unsetEnvVars(s.Suite)
}

func (s *setupTestSuite) TestSucceed() {
	s.Require().NoError(os.Setenv("PAYMENTS_HTTP_SERVER_PORT", "8001"))
	s.Require().NoError(os.Setenv("PAYMENTS_STRIPE_SIGNING_KEY", "test1234"))
	s.Require().NoError(os.Setenv("PAYMENTS_STRIPE_SECRET_KEY", "secret1234"))
	s.Require().NoError(os.Setenv("PAYMENTS_CIRCUIT_BREAKER_TIMEOUT", "10s"))
	s.Require().NoError(os.Setenv("PAYMENTS_CREDITS_SERVICE_URL", "http://localhost:8082"))
	s.Require().NoError(os.Setenv("PAYMENTS_CUSTOMERS_SERVICE_URL", "http://localhost:8083"))

	cfg, err := Setup(s.Logger)

	s.Assert().NoError(err)
	s.Assert().Equal(uint(8001), cfg.Port)
	s.Assert().Equal("test1234", cfg.Stripe.SigningKey)
	s.Assert().Equal("secret1234", cfg.Stripe.SecretKey)
	s.Assert().Equal(10*time.Second, cfg.Timeout)
	s.Assert().Equal("http://localhost:8082", cfg.CreditsURL.String())
	s.Assert().Equal("http://localhost:8083", cfg.CustomersURL.String())
}

func (s *setupTestSuite) TestDefaultValues() {
	s.Require().NoError(os.Setenv("PAYMENTS_STRIPE_SIGNING_KEY", "test1234"))
	s.Require().NoError(os.Setenv("PAYMENTS_STRIPE_SECRET_KEY", "secret1234"))
	s.Require().NoError(os.Setenv("PAYMENTS_CREDITS_SERVICE_URL", "http://localhost:8082"))
	s.Require().NoError(os.Setenv("PAYMENTS_CUSTOMERS_SERVICE_URL", "http://localhost:8083"))

	cfg, err := Setup(s.Logger)
	s.Assert().NoError(err)
	s.Assert().Equal(uint(80), cfg.Port)
	s.Assert().Equal(30*time.Second, cfg.Timeout)
	s.Assert().Equal("", cfg.Stripe.URL)
}

func (s *setupTestSuite) TestMissingEnvVars() {
	_, err := Setup(s.Logger)
	s.Assert().Error(err)
}

func (s *setupTestSuite) TestSetupWithErrors() {
	s.Require().NoError(os.Setenv("PAYMENTS_HTTP_SERVER_PORT", "ABCD"))

	_, err := Setup(s.Logger)
	s.Assert().Error(err)
}

type serverTestSuite struct {
	suite.Suite
	Logger    *log.Logger
	Config    conf.Config
	Payments  application.Service
	Credits   credits.Client
	Customers customers.Client
	Adapter   adapter.Client
}

func TestServerSuite(t *testing.T) {
	suite.Run(t, new(serverTestSuite))
}

func (s *serverTestSuite) SetupSuite() {
	var err error
	s.Require().NoError(os.Setenv("PAYMENTS_HTTP_SERVER_PORT", "8001"))
	s.Require().NoError(os.Setenv("PAYMENTS_STRIPE_SIGNING_KEY", "test1234"))
	s.Require().NoError(os.Setenv("PAYMENTS_STRIPE_SECRET_KEY", "secret1234"))
	s.Require().NoError(os.Setenv("PAYMENTS_CREDITS_SERVICE_URL", "http://localhost:8082"))
	s.Require().NoError(os.Setenv("PAYMENTS_CUSTOMERS_SERVICE_URL", "http://localhost:8083"))

	s.Logger = log.New(os.Stdout, "[TestServer] ", log.LstdFlags|log.Lshortfile|log.Lmsgprefix)
	s.Config, err = Setup(s.Logger)
	s.Require().NoError(err)
	s.Credits = fakecredits.NewClient()
	s.Customers = fakecustomers.NewClient()

	s.Adapter = adapter.NewStripeAdapter(s.Config.Stripe)

	s.Payments = application.NewPaymentsService(application.Options{
		Credits:   s.Credits,
		Customers: s.Customers,
		Adapter:   s.Adapter,
		Logger:    s.Logger,
		Timeout:   10 * time.Second,
	})
}

func (s *serverTestSuite) TestListenAndServe() {
	opts := Options{
		config:   s.Config,
		payments: s.Payments,
		logger:   s.Logger,
	}

	server := NewServer(opts)

	running := make(chan struct{})
	done := make(chan struct{})
	go func() {
		close(running)
		err := server.ListenAndServe()
		s.Require().NoError(err)
		defer close(done)
	}()

	<-running

	err := server.Shutdown(context.TODO())
	s.Assert().NoError(err)

	<-done
}

func (s *serverTestSuite) TestListenAndServeAddressInUse() {
	opts := Options{
		config:   s.Config,
		payments: s.Payments,
		logger:   s.Logger,
	}

	server := NewServer(opts)
	anotherServer := NewServer(opts)

	running := make(chan struct{})
	done := make(chan struct{})
	go func() {
		close(running)
		err := server.ListenAndServe()
		s.Require().NoError(err)
		defer close(done)
	}()

	<-running

	// Running another HTTP server listening to the same port will cause an error
	err := anotherServer.ListenAndServe()
	s.Assert().Error(err)

	// Shutting down the first server should work
	err = server.Shutdown(context.TODO())
	s.Assert().NoError(err)

	<-done
}

func (s *serverTestSuite) TestServerShutdownBeforeRunning() {
	opts := Options{
		config:   s.Config,
		payments: s.Payments,
		logger:   s.Logger,
	}

	server := NewServer(opts)

	s.Assert().NoError(server.Shutdown(context.Background()))
}

func (s *serverTestSuite) TearDownSuite() {
	unsetEnvVars(s.Suite)
}

type runTestSuite struct {
	suite.Suite
	Logger    *log.Logger
	Config    conf.Config
	Credits   credits.Client
	Payments  application.Service
	Customers customers.Client
	Adapter   adapter.Client
}

func (s *runTestSuite) SetupSuite() {
	var err error
	s.Require().NoError(os.Setenv("PAYMENTS_HTTP_SERVER_PORT", "8001"))
	s.Require().NoError(os.Setenv("PAYMENTS_STRIPE_SIGNING_KEY", "test1234"))
	s.Require().NoError(os.Setenv("PAYMENTS_STRIPE_SECRET_KEY", "secret1234"))
	s.Require().NoError(os.Setenv("PAYMENTS_CREDITS_SERVICE_URL", "http://localhost:8082"))
	s.Require().NoError(os.Setenv("PAYMENTS_CUSTOMERS_SERVICE_URL", "http://localhost:8083"))

	s.Logger = log.New(os.Stdout, "[TestRun] ", log.LstdFlags|log.Lshortfile|log.Lmsgprefix)
	s.Config, err = Setup(s.Logger)
	s.Require().NoError(err)

	s.Credits = credits.NewCreditsClientV1(s.Config.CreditsURL, s.Config.Timeout)

	s.Customers = customers.NewCustomersClientV1(s.Config.CustomersURL, s.Config.Timeout)
	s.Adapter = adapter.NewStripeAdapter(s.Config.Stripe)

	s.Payments = application.NewPaymentsService(application.Options{
		Credits:   s.Credits,
		Customers: s.Customers,
		Adapter:   s.Adapter,
		Logger:    s.Logger,
		Timeout:   10 * time.Second,
	})
}

func (s *runTestSuite) TestRunAddressInUse() {
	opts := Options{
		config: s.Config,
		logger: s.Logger,
	}

	// Run a web server
	server := NewServer(opts)

	running := make(chan struct{})
	done := make(chan struct{})
	go func() {
		close(running)
		err := server.ListenAndServe()
		s.Require().NoError(err)
		defer close(done)
	}()

	<-running

	// Running the default web server will cause an issue because there's already a server listening to the same port
	err := Run(s.Config, s.Logger)
	s.Assert().Error(err)

	// Shutting down the tmp HTTP server
	err = server.Shutdown(context.TODO())
	s.Assert().NoError(err)

	<-done
}

func (s *runTestSuite) TearDownSuite() {
	unsetEnvVars(s.Suite)
}

func TestRun(t *testing.T) {
	suite.Run(t, new(runTestSuite))
}

func unsetEnvVars(s suite.Suite) {
	s.Require().NoError(os.Unsetenv("PAYMENTS_HTTP_SERVER_PORT"))
	s.Require().NoError(os.Unsetenv("PAYMENTS_STRIPE_SIGNING_KEY"))
	s.Require().NoError(os.Unsetenv("PAYMENTS_STRIPE_SECRET_KEY"))
	s.Require().NoError(os.Unsetenv("PAYMENTS_STRIPE_URL"))
	s.Require().NoError(os.Unsetenv("PAYMENTS_CIRCUIT_BREAKER_TIMEOUT"))
	s.Require().NoError(os.Unsetenv("PAYMENTS_CREDITS_SERVICE_URL"))
	s.Require().NoError(os.Unsetenv("PAYMENTS_CUSTOMERS_SERVICE_URL"))
}
