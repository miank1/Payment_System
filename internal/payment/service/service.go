package service

import (
	"context"
	"fmt"
	"payment-system/internal/payment/model"
	"payment-system/internal/payment/repository"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo *repository.Repository
}

func New(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreatePayment(ctx context.Context, userID string, amount int64, idempotencyKey string) (*model.Payment, error) {
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

	// Step 4: Mock charge — always succeeds for now
	chargeSuccess := true

	// Step 5: Move to SUCCESS or FAILED
	if chargeSuccess {
		s.repo.UpdateStatus(ctx, id, model.StatusSuccess)
		p.Status = model.StatusSuccess
	} else {
		s.repo.UpdateStatus(ctx, id, model.StatusFailed)
		p.Status = model.StatusFailed
	}

	return p, nil
}
