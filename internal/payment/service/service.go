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
