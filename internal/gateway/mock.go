package gateway

import "log"

type MockGateway struct{}

func NewMockGateway() *MockGateway {
	return &MockGateway{}
}

func (m *MockGateway) Charge(amount int64) error {
	log.Printf("💳 Mock charging amount: %d", amount)
	return nil
}
