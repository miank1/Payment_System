package handler

import (
	"net/http"
	"payment-system/internal/payment/service"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *service.Service
}

func New(service *service.Service) *Handler {
	return &Handler{service: service}
}

type CreatePaymentRequest struct {
	UserID         string `json:"user_id" binding:"required"`
	Amount         int64  `json:"amount" binding:"required"`
	IdempotencyKey string `json:"idempotency_key" binding:"required"`
}

// CreatePayment godoc
// @Summary      Create a payment
// @Description  Create a new payment
// @Tags         payments
// @Accept       json
// @Produce      json
// @Param        payment body CreatePaymentRequest true "Payment Request"
// @Success      201 {object} model.Payment
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /payments [post]
func (h *Handler) CreatePayment(c *gin.Context) {
	var req CreatePaymentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payment, err := h.service.CreatePayment(c.Request.Context(), req.UserID, req.Amount, req.IdempotencyKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, payment)
}

// GetPayment godoc
// @Summary      Get a payment
// @Description  Get payment by ID
// @Tags         payments
// @Produce      json
// @Param        id path string true "Payment ID"
// @Success      200 {object} model.Payment
// @Failure      404 {object} map[string]string
// @Router       /payments/{id} [get]
func (h *Handler) GetPayment(c *gin.Context) {
	id := c.Param("id")

	payment, err := h.service.GetPayment(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, payment)
}

// ProcessPayment godoc
// @Summary      Process a payment
// @Description  Process payment through gateway
// @Tags         payments
// @Produce      json
// @Param        id path string true "Payment ID"
// @Success      200 {object} model.Payment
// @Failure      400 {object} map[string]string
// @Router       /payments/{id}/process [post]
func (h *Handler) ProcessPayment(c *gin.Context) {
	id := c.Param("id")

	payment, err := h.service.ProcessPayment(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, payment)
}

// HealthCheck godoc
// @Summary      Health check
// @Description  Check if system is healthy
// @Tags         health
// @Produce      json
// @Success      200 {object} map[string]string
// @Router       /healthz [get]
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now(),
	})
}
