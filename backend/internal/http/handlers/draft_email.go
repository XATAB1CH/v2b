package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/XATAB1CH/v2b/internal/http/middleware"
	"github.com/XATAB1CH/v2b/internal/services"
)

type DraftEmailHandler struct {
	svc *services.DraftEmailService
	ent *services.EntitlementsService
}

func NewDraftEmailHandler(
	svc *services.DraftEmailService,
	ent *services.EntitlementsService,
) *DraftEmailHandler {
	return &DraftEmailHandler{svc: svc, ent: ent}
}

type DraftEmailRequest struct {
	Text      string `json:"text"`
	Lang      string `json:"lang"`
	Tone      string `json:"tone"`
	Recipient string `json:"recipient"`
	Sender    string `json:"sender"`
}

type DraftEmailResponse struct {
	Email string `json:"email"`
}

func (h *DraftEmailHandler) DraftEmail(c *gin.Context) {
	// --- auth context ---
	userID := c.GetString(middleware.CtxUserIDKey)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// --- parse request ---
	var req DraftEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad json"})
		return
	}
	if len(req.Text) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "text too short"})
		return
	}

	// --- gate: subscription / free attempts ---
	attemptsLeft, allowed, reason, err := h.ent.ConsumeAttemptOrAllow(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !allowed {
		// PAYWALL
		c.JSON(http.StatusPaymentRequired, gin.H{
			"error":         "payment_required",
			"reason":        reason,       // "free_limit_reached"
			"attempts_left": attemptsLeft, // обычно 0
			"action":        "subscribe",  // для UI
		})
		return
	}

	// --- OpenAI ---
	email, err := h.svc.Draft(c.Request.Context(), services.DraftEmailParams{
		Text:      req.Text,
		Lang:      req.Lang,
		Tone:      req.Tone,
		Recipient: req.Recipient,
		Sender:    req.Sender,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// --- success ---
	c.JSON(http.StatusOK, DraftEmailResponse{
		Email: email,
	})
}
