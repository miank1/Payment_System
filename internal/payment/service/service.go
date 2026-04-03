package service

import (
	"context"
	"fmt"
	"log"
	"payment-system/internal/gateway"
	"payment-system/internal/payment/model"
	"payment-system/internal/payment/repository"
	"payment-system/pkg/queue"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo      *repository.Repository
	publisher *queue.Publisher
	gateway   gateway.PaymentGateway // interface added
}

func New(repo *repository.Repository, publisher *queue.Publisher, gw gateway.PaymentGateway) *Service {
	return &Service{repo: repo, publisher: publisher, gateway: gw}
}

func (s *Service) CreatePayment(ctx context.Context, userID string, amount int64, idempotencyKey string) (*model.Payment, error) {

	// Step 1: Check if payment already exists
	existing, err := s.repo.GetByIdempotencyKey(ctx, idempotencyKey)
	if err != nil {
		return nil, fmt.Errorf("failed to check idempotency: %w", err)
	}

	// Step 2: Already exists → return it
	if existing != nil {
		return existing, nil
	}

	p := &model.Payment{
		ID:             uuid.NewString(),
		UserID:         userID,
		Amount:         amount,
		Status:         model.StatusCreated,
		IdempotencyKey: idempotencyKey,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// Push to queue
	if err := s.publisher.Publish(ctx, p.ID); err != nil {
		return nil, fmt.Errorf("failed to publish payment: %w", err)
	}

	log.Printf("📬 Payment %s pushed to queue", p.ID)

	return p, nil
}

func (s *Service) GetPayment(ctx context.Context, id string) (*model.Payment, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}
	if p == nil {
		return nil, fmt.Errorf("payment not found")
	}
	return p, nil
}

func (s *Service) ProcessPayment(ctx context.Context, id string) (*model.Payment, error) {
	// Step 1: Fetch payment
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch payment: %w", err)
	}
	if p == nil {
		return nil, fmt.Errorf("payment not found")
	}

	// Step 2: Only CREATED payments can be processed
	if p.Status != model.StatusCreated {
		return nil, fmt.Errorf("payment is already %s", p.Status)
	}

	// Step 3: Move to PROCESSING
	if err := s.repo.UpdateStatus(ctx, id, model.StatusProcessing); err != nil {
		return nil, fmt.Errorf("failed to update status: %w", err)
	}

	// ← call gateway, check error
	if err := s.gateway.Charge(p.Amount); err != nil {
		s.repo.UpdateStatus(ctx, id, model.StatusFailed)
		p.Status = model.StatusFailed
		return p, nil
	}

	// ← no chargeSuccess variable needed
	s.repo.UpdateStatus(ctx, id, model.StatusSuccess)
	p.Status = model.StatusSuccess
	return p, nil

}
