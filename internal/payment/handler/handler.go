package handler

import (
	"net/http"
	"payment-system/internal/payment/service"

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

func (h *Handler) GetPayment(c *gin.Context) {
	id := c.Param("id")

	payment, err := h.service.GetPayment(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, payment)
}

func (h *Handler) ProcessPayment(c *gin.Context) {
	id := c.Param("id")

	payment, err := h.service.ProcessPayment(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, payment)
}
