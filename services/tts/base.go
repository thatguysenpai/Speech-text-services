package tts

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// Voice type (replace with enum if needed)
type Voice string

// Endpoint config struct
type Endpoint struct {
	URL      string `json:"url"`
	Response string `json:"response"`
}

// Main TTS function
func TTS(text string, voice Voice, outputFilePath string, playSound bool, logger *log.Logger) error {
	// Validate args
	if err := validateArgs(text, voice); err != nil {
		return err
	}

	// Load endpoints
	endpoints, err := loadEndpoints()
	if err != nil {
		logger.Printf("An error as occured, err: %v", err)
		return err
	}

	var success bool
	for _, endpoint := range endpoints {
		audioBytes, err := fetchAudioBytes(endpoint, text, voice, logger)
		if err == nil && audioBytes != nil {
			if err := saveAudioFile(outputFilePath, audioBytes); err != nil {
				logger.Printf("An error as occured, err: %v", err)
				return err
			}

			// Optionally play sound
			if playSound {
				// Example: exec.Command("mpg123", outputFilePath).Run()
			}

			success = true
			break
		}
	}

	if !success {
		return errors.New("failed to generate audio")
	}
	return nil
}

// Save audio file
func saveAudioFile(outputFilePath string, audioBytes []byte) error {
	if _, err := os.Stat(outputFilePath); err == nil {
		os.Remove(outputFilePath)
	}
	return os.WriteFile(outputFilePath, audioBytes, 0644)
}

// Fetch audio bytes
func fetchAudioBytes(endpoint Endpoint, text string, voice Voice, logger *log.Logger) ([]byte, error) {
	textChunks := splitText(text)
	audioChunks := make([]string, len(textChunks))
	var wg sync.WaitGroup
	var mu sync.Mutex
	var failed bool

	for i, chunk := range textChunks {
		wg.Add(1)
		go func(i int, chunk string) {
			defer wg.Done()
			reqBody, _ := json.Marshal(map[string]string{
				"text":  chunk,
				"voice": string(voice),
			})

			resp, err := http.Post(endpoint.URL, "application/json", bytes.NewBuffer(reqBody))
			if err != nil {
				logger.Printf("An error as occured, err: %v", err)
				failed = true
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				failed = true
				return
			}

			var result map[string]any
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				logger.Printf("An error as occured, err: %v", err)
				failed = true
				return
			}

			mu.Lock()
			audioChunks[i] = result[endpoint.Response].(string)
			mu.Unlock()
		}(i, chunk)
	}
	wg.Wait()

	if failed {
		return nil, errors.New("failed to fetch some chunks")
	}

	// Concatenate & decode base64
	fullBase64 := strings.Join(audioChunks, "")
	return base64.StdEncoding.DecodeString(fullBase64)
}

// Load endpoints from JSON
func loadEndpoints() ([]Endpoint, error) {
	execPath, _ := os.Getwd()
	jsonFilePath := filepath.Join(execPath, "internal/config", "config.json")
	file, err := os.Open(jsonFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var endpoints []Endpoint
	if err := json.NewDecoder(file).Decode(&endpoints); err != nil {
		return nil, err
	}
	return endpoints, nil
}

// Validate args
func validateArgs(text string, voice Voice) error {
	if voice == "" {
		return errors.New("'voice' must not be empty")
	}
	if text == "" {
		return errors.New("'text' must not be empty")
	}
	return nil
}

// Split text into chunks of <= 300 chars
func splitText(text string) []string {
	re := regexp.MustCompile(`.*?[.,!?:;-]|.+`)
	separatedChunks := re.FindAllString(text, -1)

	var mergedChunks []string
	var currentChunk string
	charLimit := 300

	for _, chunk := range separatedChunks {
		if len(currentChunk)+len(chunk) <= charLimit {
			currentChunk += chunk
		} else {
			if currentChunk != "" {
				mergedChunks = append(mergedChunks, currentChunk)
			}
			currentChunk = chunk
		}
	}
	if currentChunk != "" {
		mergedChunks = append(mergedChunks, currentChunk)
	}
	return mergedChunks
}
