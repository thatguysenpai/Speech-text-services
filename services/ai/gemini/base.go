package gemini

import (
	"context"
	"fmt"

	"google.golang.org/genai"

	"sts/internal/config"
)

// SendPrompt sends a prompt to Gemini with optional system instructions
func SendPrompt(systemPrompt, userPrompt string) (string, error) {
	ctx := context.Background()

	config := genai.ClientConfig{
		APIKey: config.GEMINI_API_KEY,
	}

	client, err := genai.NewClient(ctx, &config)
	if err != nil {
		return "", fmt.Errorf("failed to create Gemini client: %w", err)
	}

	model := client.GenerativeModel("gemini-1.5-flash")

	if systemPrompt != "" {
		model.SystemInstruction = &genai.Content{
			Parts: []genai.Part{genai.Text(systemPrompt)},
		}
	}

	resp, err := model.GenerateContent(ctx,
		genai.Text(userPrompt),
	)
	if err != nil {
		return "", fmt.Errorf("generation failed: %w", err)
	}

	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		if textPart, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
			return string(textPart), nil
		}
	}

	return "", fmt.Errorf("no text response found")
}
