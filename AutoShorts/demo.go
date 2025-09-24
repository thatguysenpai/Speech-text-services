package autoshorts

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

type YouTubeSegmentDownloader struct {
	APIKey     string
	Client     *openai.Client
	OutputPath string
}

// Constructor with explicit API key
func NewYouTubeSegmentDownloader(apiKey string) *YouTubeSegmentDownloader {
	client := openai.NewClient(apiKey)
	return &YouTubeSegmentDownloader{APIKey: apiKey, Client: client}
}

// Constructor from .env (mirrors uploads.go)
func LoadDownloaderFromEnv() *YouTubeSegmentDownloader {
	if err := godotenv.Load(); err != nil {
		log.Println("[WARN] No .env file found, relying on system environment")
	}
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("[ERROR] OPENAI_API_KEY not set in environment")
	}
	return NewYouTubeSegmentDownloader(apiKey)
}

// Splits transcript text into segments
func (y *YouTubeSegmentDownloader) SegmentTranscript(fullTranscript string, maxLength int) []string {
	words := strings.Fields(fullTranscript)
	var segments []string
	var current []string

	for _, word := range words {
		current = append(current, word)
		if len(current) >= maxLength {
			segments = append(segments, strings.Join(current, " "))
			current = []string{}
		}
	}
	if len(current) > 0 {
		segments = append(segments, strings.Join(current, " "))
	}
	return segments
}

// Downloads transcript using youtube-transcript-api CLI wrapper
func (y *YouTubeSegmentDownloader) DownloadTranscript(videoID string) (string, error) {
	transcriptFile := fmt.Sprintf("%s_transcript.srt", videoID)
	cmd := exec.Command("python3", "-m", "youtube_transcript_api", "--format", "srt", videoID)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to fetch transcript: %w", err)
	}

	if err := os.WriteFile(transcriptFile, out, 0644); err != nil {
		return "", err
	}
	log.Printf("[INFO] Transcript saved to %s\n", transcriptFile)
	return transcriptFile, nil
}

// Core logic to download + crop + title
func (y *YouTubeSegmentDownloader) DownloadYouTubeSegmentFromChatCompletion(youtubeURL, transcriptFile string) error {
	ctx := context.Background()

	data, err := os.ReadFile(transcriptFile)
	if err != nil {
		return err
	}
	segments := y.SegmentTranscript(string(data), 500)
	if len(segments) == 0 {
		return fmt.Errorf("no transcript segments found")
	}

	randomSegment := segments[rand.Intn(len(segments))]
	log.Println("[INFO] Selected transcript segment...")

	// Ask OpenAI for start-end and title
	resp, err := y.Client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: "gpt-4o-mini",
			Messages: []openai.ChatCompletionMessage{
				{
					Role: "system",
					Content: `Pick a 25–30s viral-worthy video segment. 
Return (hh:mm:ss,ms --> hh:mm:ss,ms) {Title}`,
				},
				{Role: "user", Content: randomSegment},
			},
			MaxTokens: 100,
		},
	)
	if err != nil {
		return err
	}
	segmentStr := resp.Choices[0].Message.Content
	log.Printf("[INFO] AI response: %s\n", segmentStr)

	// Regex for times and title
	timePattern := regexp.MustCompile(`\((\d{2}:\d{2}:\d{2},\d{3}) --> (\d{2}:\d{2}:\d{2},\d{3})\)`)
	titlePattern := regexp.MustCompile(`\{(.+?)\}`)

	timeMatch := timePattern.FindStringSubmatch(segmentStr)
	titleMatch := titlePattern.FindStringSubmatch(segmentStr)

	if len(timeMatch) < 3 || len(titleMatch) < 2 {
		return fmt.Errorf("invalid AI response: %s", segmentStr)
	}

	title := titleMatch[1]
	safeTitle := strings.ReplaceAll(regexp.MustCompile(`[<>:"/\\|?*]`).ReplaceAllString(title, ""), " ", "_")
	outputFile := fmt.Sprintf("%s.mp4", safeTitle)

	// Download video via yt-dlp
	log.Println("[INFO] Downloading video...")
	cmd := exec.Command("yt-dlp", "-f", "mp4", "-o", "downloaded.mp4", youtubeURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		_ = os.Remove("downloaded.mp4") // cleanup partial
		return fmt.Errorf("yt-dlp failed: %w", err)
	}
	videoPath := "downloaded.mp4"

	// Crop with ffmpeg to 9:16
	log.Println("[INFO] Cropping to 9:16...")
	cmd = exec.Command("ffmpeg", "-i", videoPath,
		"-vf", "crop=(in_h*9/16):in_h,scale=1080:1920",
		"-c:a", "copy", outputFile,
		"-y",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		_ = os.Remove(outputFile) // cleanup failed file
		return fmt.Errorf("ffmpeg failed: %w", err)
	}

	y.OutputPath = outputFile
	log.Printf("[INFO] Segment saved as %s\n", outputFile)

	// Use Go-native captions.go instead of captions.py
	log.Println("[INFO] Adding captions...")
	if err := ProcessVideo(y.APIKey, outputFile); err != nil {
		log.Printf("[WARN] Captioning failed: %v", err)
	}

	return nil
}
