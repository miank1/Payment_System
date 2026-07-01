package handler

import (
	"net/http"
	"payment_service/internal/service"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	Svc service.PaymentService
}

func NewPaymentHandler(svc service.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		Svc: svc,
	}
}

type CreatePaymentRequest struct {
	OrderID string  `json:"order_id" binding:"required"`
	UserID  string  `json:"user_id" binding:"required"`
	Amount  float64 `json:"amount" binding:"required,gt=0"`
}

type UpdatePaymentStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// POST /api/v1/payments
func (h *PaymentHandler) CreatePayment(c *gin.Context) {

	var req CreatePaymentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	payment, err := h.Svc.CreatePayment(
		req.OrderID,
		req.UserID,
		req.Amount,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "payment created",
		"payment": payment,
	})
}

// GET /api/v1/payments/:id
func (h *PaymentHandler) GetPayment(c *gin.Context) {

	id := c.Param("id")

	payment, err := h.Svc.GetPayment(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"payment": payment,
	})
}

// PATCH /api/v1/payments/:id/status
func (h *PaymentHandler) UpdatePaymentStatus(c *gin.Context) {

	id := c.Param("id")

	var req UpdatePaymentStatusRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	payment, err := h.Svc.UpdatePaymentStatus(
		id,
		req.Status,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "payment status updated",
		"payment": payment,
	})
}
