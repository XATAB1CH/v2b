package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/XATAB1CH/v2b/internal/config"
	"github.com/XATAB1CH/v2b/internal/services"
)

type BillingHandler struct {
	cfg config.Config
	svc *services.BillingService
}

func NewBillingHandler(cfg config.Config, svc *services.BillingService) *BillingHandler {
	return &BillingHandler{cfg: cfg, svc: svc}
}

func (h *BillingHandler) CreatePayment(c *gin.Context) {
	// пока цена/описание фиксируем здесь (потом в config)
	amount := "199.00"
	currency := "RUB"
	returnURL := "http://localhost:8080/return"
	description := "Доступ на 30 дней"

	// userID берём из middleware.AuthRequired, поэтому handler должен быть под auth group
	userID := c.GetString("user_id")

	res, err := h.svc.Create(c.Request.Context(), userID, amount, currency, returnURL, description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// DEV-only endpoint: имитация успешной оплаты
func (h *BillingHandler) MarkSucceeded(c *gin.Context) {
	if h.cfg.Env != "local" {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	paymentID := c.Param("id")
	if paymentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing payment id"})
		return
	}

	if err := h.svc.MarkSucceeded(c.Request.Context(), paymentID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
