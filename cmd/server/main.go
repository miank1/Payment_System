package main

import (
	"log"
	"payment-system/internal/payment/handler"
	"payment-system/internal/payment/repository"
	"payment-system/internal/payment/service"
	"payment-system/pkg/db"

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

	// Wire layers
	repo := repository.New(database)
	svc := service.New(repo)
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
