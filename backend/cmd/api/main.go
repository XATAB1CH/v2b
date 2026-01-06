package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/XATAB1CH/v2b/internal/clients/mailer"
	"github.com/XATAB1CH/v2b/internal/clients/openai"
	"github.com/XATAB1CH/v2b/internal/clients/stubpay"
	"github.com/XATAB1CH/v2b/internal/config"
	"github.com/XATAB1CH/v2b/internal/http/handlers"
	"github.com/XATAB1CH/v2b/internal/http/router"
	"github.com/XATAB1CH/v2b/internal/services"
	"github.com/XATAB1CH/v2b/internal/store/memory"
	"github.com/XATAB1CH/v2b/internal/store/postgres"
)

func main() {
	cfg := config.Load()

	// __________ Mailer ________
	var m mailer.Mailer
	if cfg.SMTPHost != "" && cfg.SMTPFromEmail != "" {
		m = mailer.NewSMTPMailer(mailer.SMTPConfig{
			Host:      cfg.SMTPHost,
			Port:      cfg.SMTPPort,
			Username:  cfg.SMTPUsername,
			Password:  cfg.SMTPPassword,
			FromEmail: cfg.SMTPFromEmail,
			FromName:  cfg.SMTPFromName,
			UseTLS:    cfg.SMTPUseTLS,
		})
	}

	// ---------- HTTP client ----------
	httpClient := &http.Client{Timeout: cfg.HTTPTimeout}

	// ---------- OpenAI ----------
	openaiClient := openai.NewClient(cfg.OpenAIKey, cfg.OpenAIBase, cfg.OpenAIModel, httpClient)

	// ---------- Postgres (pgxpool) ----------
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("postgres connection failed: %v", err)
	}
	defer pool.Close()

	usersRepo := postgres.NewUsersRepo(pool)
	entRepo := postgres.NewEntitlementsRepo(pool)

	// Адаптер для BillingService (интерфейс EntitlementsStore из domain/billing)
	entStoreAdapter := postgres.NewEntitlementsStoreAdapter(entRepo)

	// ---------- In-memory stores ----------
	otpStore := memory.NewOTPStore()
	paymentsStore := memory.NewPaymentsStore()

	// ---------- Services ----------
	authSvc := services.NewAuthService(
		otpStore,
		usersRepo,
		entRepo,
		cfg.JWTAccessSecret,
		cfg.JWTAccessTTLMins,
		cfg.OTPTTLMins,
		cfg.OTPCodeLength,
		cfg.OTPMaxAttempts,
		cfg.FreeAttemptsLimit,
		m,
		cfg.Env,
	)

	entSvc := services.NewEntitlementsService(entRepo)

	// Stub payment provider (позже заменим на YooKassa)
	// public base url пока хардкодом для локалки
	publicBaseURL := "http://localhost:" + cfg.Port
	stubProvider := stubpay.New(publicBaseURL)

	billingSvc := services.NewBillingService(
		stubProvider,
		paymentsStore,
		entStoreAdapter,
		cfg.SubscriptionDurationDays,
	)

	draftSvc := services.NewDraftEmailService(openaiClient)

	// ---------- Handlers ----------
	authHandler := handlers.NewAuthHandler(authSvc)
	meHandler := handlers.NewMeHandler(entSvc)
	billingHandler := handlers.NewBillingHandler(cfg, billingSvc)

	// Draft handler должен уметь делать gate через EntitlementsService
	draftHandler := handlers.NewDraftEmailHandler(draftSvc, entSvc)

	// ---------- Router ----------
	r := router.New(cfg, router.Handlers{
		DraftEmail: draftHandler,
		Auth:       authHandler,
		Me:         meHandler,
		Billing:    billingHandler,
	})

	addr := ":" + cfg.Port
	log.Printf("listening on %s", addr)
	_ = r.Run(addr)
}
