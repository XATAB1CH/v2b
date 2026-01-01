package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"

	"github.com/XATAB1CH/v2b/internal/clients/openai"
	"github.com/XATAB1CH/v2b/internal/config"
	"github.com/XATAB1CH/v2b/internal/http/handlers"
	"github.com/XATAB1CH/v2b/internal/http/router"
	"github.com/XATAB1CH/v2b/internal/services"
)

func main() {
	_ = godotenv.Load("../.env")

	cfg := config.Load()

	httpClient := &http.Client{
		Timeout: cfg.HTTPTimeout,
	}

	openaiClient := openai.NewClient(cfg.OpenAIKey, cfg.OpenAIBase, cfg.OpenAIModel, httpClient)

	draftSvc := services.NewDraftEmailService(openaiClient)
	draftHandler := handlers.NewDraftEmailHandler(draftSvc)

	r := router.New(cfg, router.Handlers{
		DraftEmail: draftHandler,
	})

	addr := ":" + cfg.Port
	log.Printf("listening on %s", addr)
	_ = r.Run(addr)
}
