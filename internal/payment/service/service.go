package service

import (
	"context"
	"fmt"
	"log"
	"payment-system/internal/gateway"
	"payment-system/internal/ledger"
	"payment-system/internal/notifier"
	"payment-system/internal/payment/model"
	"payment-system/internal/payment/repository"
	"payment-system/pkg/logger"
	"payment-system/pkg/queue"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Service struct {
	repo      *repository.Repository
	publisher *queue.Publisher
	gateway   gateway.PaymentGateway
	ledger    *ledger.Repository
	notifier  notifier.Notifier
}

// ← notifier added to New()
func New(
	repo *repository.Repository,
	publisher *queue.Publisher,
	gw gateway.PaymentGateway,
	ledger *ledger.Repository,
	notifier notifier.Notifier,
) *Service {
	return &Service{
		repo:      repo,
		publisher: publisher,
		gateway:   gw,
		ledger:    ledger,
		notifier:  notifier,
	}
}

func (s *Service) CreatePayment(ctx context.Context, userID string, amount int64, idempotencyKey string) (*model.Payment, error) {
	existing, err := s.repo.GetByIdempotencyKey(ctx, idempotencyKey)
	if err != nil {
		return nil, fmt.Errorf("failed to check idempotency: %w", err)
	}
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

	if err := s.publisher.Publish(ctx, p.ID); err != nil {
		return nil, fmt.Errorf("failed to publish payment: %w", err)
	}

	logger.Log.Info("payment pushed to queue",
		zap.String("payment_id", p.ID),
		zap.Int64("amount", amount),
	)
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

	// Step 1: Begin transaction
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Step 2: Lock row with SELECT FOR UPDATE
	p, err := s.repo.GetByIDForUpdate(ctx, tx, id)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to fetch payment: %w", err)
	}
	if p == nil {
		tx.Rollback()
		return nil, fmt.Errorf("payment not found")
	}

	// Step 3: Already processed → skip
	if p.Status != model.StatusCreated {
		tx.Rollback()
		log.Printf("⚠️  Payment %s already %s, skipping", id, p.Status)
		return p, nil
	}

	// Step 4: Move to PROCESSING inside transaction
	if err := s.repo.UpdateStatusTx(ctx, tx, id, model.StatusProcessing); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update status: %w", err)
	}

	// Step 5: Commit PROCESSING status
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Step 6: Charge via gateway
	if err := s.gateway.Charge(p.Amount); err != nil {
		s.repo.UpdateStatus(ctx, id, model.StatusFailed)
		p.Status = model.StatusFailed
		s.notifier.Send(id, model.StatusFailed, p.Amount)
		log.Printf("🔔 Notification sent → FAILED for payment %s", id)
		return p, nil
	}

	// Step 7: Move to SUCCESS
	s.repo.UpdateStatus(ctx, id, model.StatusSuccess)
	p.Status = model.StatusSuccess

	// Step 8: Write ledger
	s.ledger.AddEntry(ctx, id, ledger.TypeDebit, p.Amount)
	s.ledger.AddEntry(ctx, id, ledger.TypeCredit, p.Amount)
	logger.Log.Info("ledger entries written",
		zap.String("payment_id", id),
	)

	// Step 9: Notify
	s.notifier.Send(id, model.StatusSuccess, p.Amount)
	logger.Log.Info("notification sent",
		zap.String("payment_id", id),
		zap.String("status", model.StatusSuccess),
	)

	return p, nil
}
