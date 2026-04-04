package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"payment-system/internal/gateway"
	"payment-system/internal/ledger"
	"payment-system/internal/notifier"
	"payment-system/internal/payment/handler"
	"payment-system/internal/payment/repository"
	"payment-system/internal/payment/service"
	"payment-system/pkg/db"
	"payment-system/pkg/logger"
	"payment-system/pkg/queue"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

const maxRetries = 3

func main() {

	// Init logger
	logger.Init()
	defer logger.Sync()

	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Connect DB
	database := db.Connect()
	defer database.Close()

	// Connect Queue
	publisher, err := queue.NewPublisher(
		os.Getenv("RABBITMQ_URL"),
		os.Getenv("QUEUE_NAME"),
	)

	if err != nil {
		log.Fatalf("failed to connect queue: %v", err)
	}
	defer publisher.Close()

	// Connect Consumer
	consumer, err := queue.NewConsumer(
		os.Getenv("RABBITMQ_URL"),
		os.Getenv("QUEUE_NAME"),
	)
	if err != nil {
		log.Fatalf("failed to connect consumer: %v", err)
	}
	defer consumer.Close()

	// Wire layers
	ledgerRepo := ledger.NewRepository(database)
	repo := repository.New(database)
	gw := gateway.NewMockGateway()
	ntf := notifier.NewMockNotifier()
	svc := service.New(repo, publisher, gw, ledgerRepo, ntf)
	h := handler.New(svc)

	// Start worker in background goroutine
	go startWorker(consumer, publisher, svc)

	// Setup routes
	r := gin.Default()

	r.POST("/payments", h.CreatePayment)
	r.GET("/payments/:id", h.GetPayment)
	r.POST("/payments/:id/process", h.ProcessPayment)
	r.GET("/healthz", h.HealthCheck)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Println("🚀 Server running on :8080")
		r.Run(":8080")
	}()

	<-quit
	log.Println("🛑 Shutting down...")
}

func startWorker(consumer *queue.Consumer, publisher *queue.Publisher, svc *service.Service) {
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

		for attempt := 1; attempt <= maxRetries; attempt++ {
			_, err := svc.ProcessPayment(context.Background(), paymentID)
			if err == nil {
				success = true
				break
			}

			lastErr = err
			log.Printf("⚠️  Attempt %d failed: %v", attempt, err)

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
			publisher.PublishToDLQ(context.Background(), paymentID)
			log.Printf("📨 Payment %s sent to DLQ", paymentID)
			msg.Nack(false, false)
		}
	}
}
