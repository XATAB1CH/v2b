package services

import (
	"context"
)

type LLM interface {
	Generate(ctx context.Context, prompt string) (string, error)
}

type DraftEmailService struct {
	llm LLM
}

func NewDraftEmailService(llm LLM) *DraftEmailService {
	return &DraftEmailService{llm: llm}
}

func (s *DraftEmailService) Draft(ctx context.Context, p DraftEmailParams) (string, error) {
	prompt := BuildDraftEmailPrompt(p)
	return s.llm.Generate(ctx, prompt)
}
