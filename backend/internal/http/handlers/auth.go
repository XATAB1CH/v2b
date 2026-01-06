package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/XATAB1CH/v2b/internal/services"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	svc *services.AuthService
}

func NewAuthHandler(svc *services.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

type OTPRequest struct {
	Email string `json:"email" binding:"required,email"`
}

func (h *AuthHandler) RequestOTP(c *gin.Context) {
	// Читаем raw body через Gin (он читает тело корректно)
	raw, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "read body: " + err.Error()})
		return
	}
	if len(raw) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty body"})
		return
	}

	// Парсим JSON вручную
	var req OTPRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "bad json: " + err.Error(),
			"raw":   string(raw), // можно убрать позже
		})
		return
	}

	if strings.TrimSpace(req.Email) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}

	_, err = h.svc.RequestOTP(req.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type OTPVerifyRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required"`
}

func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req OTPVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tok, err := h.svc.VerifyOTP(c.Request.Context(), req.Email, req.Code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid code"})
		return
	}

	c.JSON(http.StatusOK, tok)
}
