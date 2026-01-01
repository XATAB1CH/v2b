package router

import (
	"github.com/XATAB1CH/v2b/internal/config"
	"github.com/XATAB1CH/v2b/internal/http/handlers"
	"github.com/XATAB1CH/v2b/internal/http/middleware"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	DraftEmail *handlers.DraftEmailHandler
}

func New(cfg config.Config, h Handlers) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS(cfg))

	api := r.Group("/api")
	{
		api.POST("/draft-email", h.DraftEmail.DraftEmail)
	}

	return r
}
