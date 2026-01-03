package router

import (
	"github.com/XATAB1CH/v2b/internal/config"
	"github.com/XATAB1CH/v2b/internal/http/handlers"
	"github.com/XATAB1CH/v2b/internal/http/middleware"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	DraftEmail *handlers.DraftEmailHandler
	Auth       *handlers.AuthHandler
	Me         *handlers.MeHandler
	Billing    *handlers.BillingHandler
}

func New(cfg config.Config, h Handlers) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS(cfg))

	api := r.Group("/api")

	// --- Public auth ---
	api.POST("/auth/otp/request", h.Auth.RequestOTP)
	api.POST("/auth/otp/verify", h.Auth.VerifyOTP)

	// --- Private (JWT) ---
	private := api.Group("")
	private.Use(middleware.AuthRequired(cfg))
	{
		private.GET("/me", h.Me.GetMe)
		private.POST("/draft-email", h.DraftEmail.DraftEmail)
		private.POST("/billing/payments", h.Billing.CreatePayment)
	}

	// --- Dev-only endpoints ---
	dev := api.Group("/dev")
	dev.Use(middleware.AuthRequired(cfg))
	{
		dev.POST("/payments/:id/mark-succeeded", h.Billing.MarkSucceeded)
	}

	return r
}
