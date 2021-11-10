package fake

import (
	"github.com/stretchr/testify/mock"
	customers "gitlab.com/ignitionrobotics/billing/customers/pkg/api"
	"gitlab.com/ignitionrobotics/billing/payments/pkg/adapter"
	"gitlab.com/ignitionrobotics/billing/payments/pkg/api"
)

var _ adapter.Client = (*Adapter)(nil)

// Adapter is a fake implementation of adapter.Client.
type Adapter struct {
	mock.Mock
}

// CreateCustomer mocks a CreateCustomer call.
func (a *Adapter) CreateCustomer(application, handle string) (string, error) {
	args := a.Called(application, handle)
	return args.String(0), args.Error(1)
}

// CreateSession mocks a CreateSession call.
func (a *Adapter) CreateSession(req api.CreateSessionRequest, cus customers.CustomerResponse) (api.CreateSessionResponse, error) {
	args := a.Called(req, cus)
	res := args.Get(0).(api.CreateSessionResponse)
	return res, args.Error(1)
}

// GenerateChargeRequest mocks a GenerateChargeRequest call.
func (a *Adapter) GenerateChargeRequest(body []byte, params map[string][]string) (api.ChargeRequest, error) {
	args := a.Called(body, params)
	res := args.Get(0).(api.ChargeRequest)
	return res, args.Error(1)
}
