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
	"payment-system/pkg/queue"
	"time"

	"github.com/google/uuid"
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

	// Step 2: Only CREATED can be processed
	if p.Status != model.StatusCreated {
		return nil, fmt.Errorf("payment is already %s", p.Status)
	}

	// Step 3: Move to PROCESSING
	if err := s.repo.UpdateStatus(ctx, id, model.StatusProcessing); err != nil {
		return nil, fmt.Errorf("failed to update status: %w", err)
	}

	// Step 4: Charge via gateway
	if err := s.gateway.Charge(p.Amount); err != nil {
		s.repo.UpdateStatus(ctx, id, model.StatusFailed)
		p.Status = model.StatusFailed

		// ← notify on failure too
		s.notifier.Send(id, model.StatusFailed, p.Amount)
		log.Printf("🔔 Notification sent → FAILED for payment %s", id)

		return p, nil
	}

	// Step 5: Move to SUCCESS
	s.repo.UpdateStatus(ctx, id, model.StatusSuccess)
	p.Status = model.StatusSuccess

	// Step 6: Write ledger
	s.ledger.AddEntry(ctx, id, ledger.TypeDebit, p.Amount)
	s.ledger.AddEntry(ctx, id, ledger.TypeCredit, p.Amount)
	log.Printf("📒 Ledger entries written for payment %s", id)

	// Step 7: Notify success
	s.notifier.Send(id, model.StatusSuccess, p.Amount)
	log.Printf("🔔 Notification sent → SUCCESS for payment %s", id)

	return p, nil
}
