package openai

import "bytes"

type OpenAIResponsesRequest struct {
	Model string `json:"model"`
	Input string `json:"input"` // ВАЖНО: строкой
}

type OpenAIResponsesResponse struct {
	OutputText string `json:"output_text"`
	Output     []struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
}

func extractText(r OpenAIResponsesResponse) string {
	if r.OutputText != "" {
		return r.OutputText
	}
	// fallback: собираем текст из output[].content[].text
	var buf bytes.Buffer
	for _, out := range r.Output {
		for _, c := range out.Content {
			if c.Type == "output_text" || c.Text != "" {
				if buf.Len() > 0 {
					buf.WriteString("\n")
				}
				buf.WriteString(c.Text)
			}
		}
	}
	return buf.String()
}
