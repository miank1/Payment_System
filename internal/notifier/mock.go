package notifier

import "log"

type MockNotifier struct{}

func NewMockNotifier() *MockNotifier {
	return &MockNotifier{}
}

func (m *MockNotifier) Send(paymentID string, status string, amount int64) error {
	log.Printf("🔔 Notification → Payment %s | Status: %s | Amount: %d",
		paymentID, status, amount)
	return nil
}
