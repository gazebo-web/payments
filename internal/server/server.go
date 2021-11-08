package server

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	credits "gitlab.com/ignitionrobotics/billing/credits/pkg/client"
	customers "gitlab.com/ignitionrobotics/billing/customers/pkg/client"
	"gitlab.com/ignitionrobotics/billing/payments/internal/conf"
	"gitlab.com/ignitionrobotics/billing/payments/pkg/adapter"
	"gitlab.com/ignitionrobotics/billing/payments/pkg/application"
	"log"
	"net/http"
)

// Setup initializes the conf.Config to run the web server.
func Setup(logger *log.Logger) (conf.Config, error) {
	logger.Println("Parsing config")
	var c conf.Config
	if err := c.Parse(); err != nil {
		logger.Println("Failed to parse config:", err)
		return conf.Config{}, err
	}
	return c, nil
}

// Run runs the web server using the given config.
func Run(config conf.Config, logger *log.Logger) error {
	logger.Println("Initializing Credits HTTP client")
	creditsClient := credits.NewClient()

	logger.Println("Initializing Customers HTTP client")
	customersClient := customers.NewClient()

	logger.Println("Initializing Stripe adapter")
	stripeAdapter := adapter.NewStripeAdapter(config.Stripe)

	logger.Println("Initializing Payments service")
	ps := application.NewPaymentsService(application.Options{
		Credits:   creditsClient,
		Customers: customersClient,
		Adapter:   stripeAdapter,
		Logger:    logger,
		Timeout:   config.Timeout,
	})

	logger.Println("Initializing HTTP server")
	s := NewServer(Options{
		config:   config,
		payments: ps,
		logger:   logger,
	})

	if err := s.ListenAndServe(); err != nil {
		logger.Println("Error while running HTTP server:", err)
		return err
	}
	return nil
}

// Options contains a set of components to be used when initializing a web server.
type Options struct {
	config   conf.Config
	payments application.Service
	logger   *log.Logger
	adapter  adapter.Client
}

// Server is an HTTP web server used to expose api.PaymentsV1 endpoints. It prepares the input for each
// service operation and returns a serialized JSON response from each operation output.
type Server struct {
	// payments contains an implementation of application.Service
	payments application.Service

	// logger contains the logger used to print debug information.
	logger *log.Logger

	// router is used to route requests in the HTTP server.
	router chi.Router

	// port is the HTTP port used to listen for incoming requests.
	port uint

	// httpServer is used to serve the router with fine-grained control of ListenAndServe and Shutdown operations.
	httpServer http.Server

	// adapter is used to generate charge requests from incoming webhook events. It contains an adapter implementation
	// such as Stripe.
	adapter adapter.Client
}

// ListenAndServe starts listening in the port defined on conf.Config. It's in charge of serving the different endpoints.
func (s *Server) ListenAndServe() error {
	s.logger.Println("Listening on", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown shuts the web server down.
func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}

// getAddress returns a valid address (host:port) representation that the server will listen to.
func (s *Server) getAddress() string {
	return fmt.Sprintf(":%d", s.port)
}

// NewServer initializes a new web server that will serve api.PaymentsV1 and api.ChargerV1 methods.
func NewServer(opts Options) *Server {
	s := Server{
		payments: opts.payments,
		logger:   opts.logger,
		port:     opts.config.Port,
		adapter:  opts.adapter,
	}

	s.router = chi.NewRouter()
	s.router.HandleFunc("/stripe/webhook", s.StripeWebhook)

	s.httpServer = http.Server{
		Addr:    s.getAddress(),
		Handler: s.router,
	}
	return &s
}
