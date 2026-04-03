package main

import (
	"context"
	"log"
	"os"
	"payment-system/internal/gateway"
	"payment-system/internal/ledger"
	"payment-system/internal/payment/repository"
	"payment-system/internal/payment/service"
	"payment-system/pkg/db"
	"payment-system/pkg/queue"

	"github.com/joho/godotenv"
)

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
	repo := repository.New(database)
	publisher, _ := queue.NewPublisher(
		os.Getenv("RABBITMQ_URL"),
		os.Getenv("QUEUE_NAME"),
	)

	ledgerRepo := ledger.NewRepository(database)
	gw := gateway.NewMockGateway()
	svc := service.New(repo, publisher, gw, ledgerRepo)

	// Start consuming
	msgs, err := consumer.Consume()
	if err != nil {
		log.Fatalf("failed to consume: %v", err)
	}

	log.Println("👷 Worker running, waiting for payments...")

	for msg := range msgs {
		paymentID := string(msg.Body)
		log.Printf("⚙️  Processing payment: %s", paymentID)

		_, err := svc.ProcessPayment(context.Background(), paymentID)
		if err != nil {
			log.Printf("❌ Failed to process payment %s: %v", paymentID, err)
			msg.Nack(false, true) // requeue
			continue
		}

		log.Printf("✅ Payment %s processed successfully", paymentID)
		msg.Ack(false)
	}
}
