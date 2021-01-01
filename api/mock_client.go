package api

import (
	"errors"
)

// MockAPIClient ...
type MockAPIClient struct {
	returnError        bool
	appMetricCallsMade uint64
}

// NewMockAPICClient ...
func NewMockAPICClient(returnError bool) *MockAPIClient {
	return &MockAPIClient{returnError: returnError}
}

// SendServerMetrics ...
func (ac *MockAPIClient) SendServerMetrics(data string) (*Response, error) {
	if ac.returnError {
		return nil, errors.New("error")
	}

	return nil, nil
}

// SendAppMetrics ...
func (ac *MockAPIClient) SendAppMetrics(data string) (*Response, error) {
	if ac.returnError {
		ac.appMetricCallsMade++

		return nil, errors.New("error")
	}
	ac.appMetricCallsMade++
	return nil, nil
}
