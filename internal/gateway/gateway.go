package gateway

type PaymentGateway interface {
	Charge(amount int64) error
}
