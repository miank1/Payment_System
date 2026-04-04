package notifier

// Contract
// Anyone who wants to be a notifier
// MUST have a Send() function

type Notifier interface {
	Send(paymentID string, status string, amount int64) error
}
