package service

import (
	"errors"
	models "payment_service/internal/model"
	"payment_service/internal/repository"

	"github.com/google/uuid"
)

const (
	PaymentPending = "PENDING"
	PaymentSuccess = "SUCCESS"
	PaymentFailed  = "FAILED"
)

type PaymentService interface {
	CreatePayment(orderID, userID string, amount float64) (*models.Payment, error)
	GetPayment(id string) (*models.Payment, error)
	UpdatePaymentStatus(id, status string) (*models.Payment, error)
}

type paymentService struct {
	repo repository.PaymentRepository
}

func NewPaymentService(repo repository.PaymentRepository) PaymentService {
	return &paymentService{
		repo: repo,
	}
}

func (s *paymentService) CreatePayment(orderID, userID string, amount float64) (*models.Payment, error) {

	orderUUID, err := uuid.Parse(orderID)
	if err != nil {
		return nil, errors.New("invalid order id")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	payment := &models.Payment{
		OrderID: orderUUID,
		UserID:  userUUID,
		Amount:  amount,
		Status:  "pending",
	}

	if err := s.repo.Create(payment); err != nil {
		return nil, err
	}

	return payment, nil
}

func (s *paymentService) GetPayment(id string) (*models.Payment, error) {

	paymentID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid payment id")
	}

	return s.repo.GetByID(paymentID)
}

func (s *paymentService) UpdatePaymentStatus(
	id string,
	status string,
) (*models.Payment, error) {

	if status != PaymentPending &&
		status != PaymentSuccess &&
		status != PaymentFailed {
		return nil, errors.New("invalid payment status")
	}

	paymentID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid payment id")
	}

	payment, err := s.repo.GetByID(paymentID)
	if err != nil {
		return nil, err
	}

	payment.Status = status

	if err := s.repo.Update(payment); err != nil {
		return nil, err
	}

	return payment, nil
}
