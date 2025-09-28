package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

var GEMINI_API_KEY string

// LoadEnv loads environment variables. Use `logger` to avoid shadowing the log package.
func LoadEnv(logger *log.Logger) {
	if err := godotenv.Load(); err != nil {
		logger.Println("[WARN] No .env file found, relying on system environment")
	}

	GEMINI_API_KEY = os.Getenv("GEMINI_API_KEY")
	if GEMINI_API_KEY == "" {
		logger.Fatal("[ERROR] GEMINI_API_KEY not set in environment")
	}

	logger.Println("ENV loaded successfully")
}

// GeminiClient encapsulates calling the Gemini endpoint.
type GeminiClient struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	Logger     *log.Logger
}

// NewGeminiClient constructs a client. If apiKey == "" it falls back to GEMINI_API_KEY.
func NewGeminiClient(apiKey string, logger *log.Logger) *GeminiClient {
	if apiKey == "" {
		apiKey = GEMINI_API_KEY
	}
	return &GeminiClient{
		APIKey:     apiKey,
		BaseURL:    "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent",
		HTTPClient: &http.Client{},
		Logger:     logger,
	}
}

// Request/response shapes (match the simple "contents.parts.text" style used earlier)
type geminiRequest struct {
	Contents []struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"contents"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// Generate sends `prompt` to Gemini and returns the first text result.
func (c *GeminiClient) Generate(prompt string) (string, error) {
	// build request payload
	reqBody := geminiRequest{
		Contents: []struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		}{
			{
				Parts: []struct {
					Text string `json:"text"`
				}{
					{Text: prompt},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(reqBody); err != nil {
		return "", fmt.Errorf("encode request: %w", err)
	}

	// Use ?key=APIKEY for simple API-key auth (or change to Bearer header if needed)
	url := c.BaseURL + "?key=" + c.APIKey
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("non-200 response %d: %s", resp.StatusCode, string(body))
	}

	var gr geminiResponse
	if err := json.Unmarshal(body, &gr); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	if len(gr.Candidates) > 0 && len(gr.Candidates[0].Content.Parts) > 0 {
		return gr.Candidates[0].Content.Parts[0].Text, nil
	}

	return "", fmt.Errorf("no text found in Gemini response")
}
