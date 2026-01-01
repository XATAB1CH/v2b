package handlers

import (
	"net/http"

	"github.com/XATAB1CH/v2b/internal/services"

	"github.com/gin-gonic/gin"
)

type DraftEmailHandler struct {
	svc *services.DraftEmailService
}

func NewDraftEmailHandler(svc *services.DraftEmailService) *DraftEmailHandler {
	return &DraftEmailHandler{svc: svc}
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
	var req DraftEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad json"})
		return
	}
	if len(req.Text) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "text too short"})
		return
	}

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

	c.JSON(http.StatusOK, DraftEmailResponse{Email: email})
}
