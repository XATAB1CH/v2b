package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/XATAB1CH/v2b/internal/http/middleware"
	"github.com/XATAB1CH/v2b/internal/services"
)

type MeHandler struct {
	ent *services.EntitlementsService
}

func NewMeHandler(ent *services.EntitlementsService) *MeHandler {
	return &MeHandler{ent: ent}
}

func (h *MeHandler) GetMe(c *gin.Context) {
	userID := c.GetString(middleware.CtxUserIDKey)

	me, err := h.ent.GetMe(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, me)
}
