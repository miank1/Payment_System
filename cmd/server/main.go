package main

import (
	"log"
	"os"
	"payment-system/internal/gateway"
	"payment-system/internal/ledger"
	"payment-system/internal/notifier"
	"payment-system/internal/payment/handler"
	"payment-system/internal/payment/repository"
	"payment-system/internal/payment/service"
	"payment-system/pkg/db"
	"payment-system/pkg/queue"

	"github.com/gin-gonic/gin"
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

	// Connect Queue
	publisher, err := queue.NewPublisher(
		os.Getenv("RABBITMQ_URL"),
		os.Getenv("QUEUE_NAME"),
	)

	if err != nil {
		log.Fatalf("failed to connect queue: %v", err)
	}
	defer publisher.Close()

	// Wire layers
	ledgerRepo := ledger.NewRepository(database)
	repo := repository.New(database)
	gw := gateway.NewMockGateway()
	ntf := notifier.NewMockNotifier()
	svc := service.New(repo, publisher, gw, ledgerRepo, ntf)
	h := handler.New(svc)

	// Setup routes
	r := gin.Default()

	r.POST("/payments", h.CreatePayment)
	r.GET("/payments/:id", h.GetPayment)
	r.POST("/payments/:id/process", h.ProcessPayment)

	// Start server
	log.Println("🚀 Server running on :8080")
	r.Run(":8080")
}
