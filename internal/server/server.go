package server

import (
	"gitlab.com/ignitionrobotics/billing/payments/internal/conf"
	"log"
)

// Setup initializes the conf.Config to run the web server.
func Setup(logger *log.Logger) (conf.Config, error) {
	return conf.Config{}, nil
}

// Run runs the web server using the given config.
func Run(config conf.Config, logger *log.Logger) error {
	return nil
}
