package main

import (
	"context"
	"log"
	"os"
	"payment-system/internal/gateway"
	"payment-system/internal/ledger"
	"payment-system/internal/notifier"
	"payment-system/internal/payment/repository"
	"payment-system/internal/payment/service"
	"payment-system/pkg/db"
	"payment-system/pkg/queue"
	"time"

	"github.com/joho/godotenv"
)

const maxRetries = 3

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Connect DB
	database := db.Connect()
	defer database.Close()

	// Connect Queue Consumer
	consumer, err := queue.NewConsumer(
		os.Getenv("RABBITMQ_URL"),
		os.Getenv("QUEUE_NAME"),
	)
	if err != nil {
		log.Fatalf("failed to connect consumer: %v", err)
	}
	defer consumer.Close()

	// Wire layers
	publisher, _ := queue.NewPublisher(
		os.Getenv("RABBITMQ_URL"),
		os.Getenv("QUEUE_NAME"),
	)
	ledgerRepo := ledger.NewRepository(database)
	gw := gateway.NewMockGateway()
	repo := repository.New(database)
	ntf := notifier.NewMockNotifier()
	svc := service.New(repo, publisher, gw, ledgerRepo, ntf)

	// Start consuming
	msgs, err := consumer.Consume()
	if err != nil {
		log.Fatalf("failed to consume: %v", err)
	}

	log.Println("👷 Worker running, waiting for payments...")

	for msg := range msgs {
		paymentID := string(msg.Body)
		log.Printf("⚙️  Processing payment: %s", paymentID)

		var lastErr error
		success := false

		// Retry loop
		for attempt := 1; attempt <= maxRetries; attempt++ {
			_, err := svc.ProcessPayment(context.Background(), paymentID)
			if err == nil {
				success = true
				break
			}

			lastErr = err
			log.Printf("⚠️  Attempt %d failed for payment %s: %v", attempt, paymentID, err)

			if attempt < maxRetries {
				log.Printf("🔄 Retrying in 2 seconds...")
				time.Sleep(2 * time.Second)
			}
		}

		if success {
			log.Printf("✅ Payment %s processed successfully", paymentID)
			msg.Ack(false)
		} else {
			log.Printf("💀 Payment %s failed after %d attempts: %v", paymentID, maxRetries, lastErr)

			// Push to DLQ
			if err := publisher.PublishToDLQ(context.Background(), paymentID); err != nil {
				log.Printf("❌ Failed to push to DLQ: %v", err)
			} else {
				log.Printf("📨 Payment %s sent to DLQ", paymentID)
			}

			msg.Nack(false, false) // drop from main queue
		}
	}
}
