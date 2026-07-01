package main

import (
	"log"
	"os"
	"payment_service/internal/handler"
	models "payment_service/internal/model"
	"payment_service/internal/repository"
	"payment_service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/miank1/ecommerce_backend/pkg/config"
	"github.com/miank1/ecommerce_backend/pkg/db"
	"github.com/miank1/ecommerce_backend/pkg/logger"
	"github.com/miank1/ecommerce_backend/pkg/middleware"
)

func LoadEnv() {
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")
}

func main() {

	logger.Init()
	defer logger.Sync()

	LoadEnv()

	// Load env
	if err := godotenv.Load(".env"); err != nil {
		log.Println("⚠️ No .env file found")
	}

	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		log.Fatal("❌ DATABASE_DSN not configured")
	}

	// DB connection
	gormDB, err := db.InitDB(dsn)
	if err != nil {
		log.Fatalf("❌ Failed to connect DB: %v", err)
	}

	// Auto migrate
	if err := gormDB.AutoMigrate(&models.Payment{}); err != nil {
		log.Fatalf("❌ AutoMigrate failed: %v", err)
	}

	// Dependency Injection
	paymentRepo := repository.NewPaymentRepository(gormDB)
	paymentService := service.NewPaymentService(paymentRepo)
	paymentHandler := handler.NewPaymentHandler(paymentService)

	// Gin
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	// Health
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "paymentservice up",
		})
	})

	api := r.Group("/payments")

	api.Use(middleware.JWTAuth())
	{
		api.GET("/:id", paymentHandler.GetPayment)
		api.POST("", paymentHandler.CreatePayment)
		api.PATCH("/:id/status", paymentHandler.UpdatePaymentStatus)
	}

	port := config.GetEnv("PORT", "8085")

	log.Printf("🚀 PaymentService running on port %s", port)

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
