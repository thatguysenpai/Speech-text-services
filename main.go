package main

import (
	"log"
	"sts/internal/models"
	"sts/services/stt"
	"sts/services/tts"
	"sts/utils"
	"sts/services/captions"
    "fmt"
    "log"
    "time"

    "sts/services/upload"
    "sts/utils"
)


	// MakeShorts generates and uploads YouTube shorts automatically
func MakeShorts() {
	videosGoal := 30
	videosPerDayGoal := 2
	scheduledTimeHour1 := 6
	amPM := "AM"
	videoID := "c8VcUnz3nVc"
	scheduledIncrement := 9

	scheduledTimeHour := scheduledTimeHour1
	today := time.Now()
	targetDate := today.AddDate(0, 0, 1)
	videosUploaded := 0
	videosDayUploaded := 0

	for videosUploaded < videosGoal {
		if videosDayUploaded >= videosPerDayGoal {
			videosDayUploaded = 0
			targetDate = targetDate.AddDate(0, 0, 1)
			scheduledTimeHour = scheduledTimeHour1
			amPM = "AM"
		}

		dateStr := targetDate.Format("Jan 02, 2006")
			
		if scheduledTimeHour >= 12 {
			if amPM == "AM" {
				amPM = "PM"
			} else {
				amPM = "AM"
				targetDate = targetDate.AddDate(0, 0, 1)
				dateStr = targetDate.Format("Jan 02, 2006")
			}
			scheduledTimeHour = scheduledTimeHour % 12
		}

		scheduledTime := fmt.Sprintf("%d:00 %s", scheduledTimeHour, amPM)

		fmt.Println("----------")
		fmt.Println(dateStr)
		fmt.Println(scheduledTime)
		fmt.Println("----------")

			// === Downloader logic ===
		youtubeURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
		downloader := NewYouTubeSegmentDownloader(openaiAPIKey)

		transcriptFilePath, err := downloader.DownloadTranscript(videoID)
		if err != nil {
			handleError(err)
			continue
		}

		err = downloader.DownloadYouTubeSegmentFromChatCompletion(youtubeURL, transcriptFilePath)
		if err != nil {
			handleError(err)
			continue
		}

		filePath := filepath.Join(projectPath, downloader.OutputPath)

		// === Uploader logic ===
		uploader := LoadUploaderFromEnv(filePath, dateStr, scheduledTime) // ✅ use uploads.go helper
		uploader.SetupDriver()

		if err := uploader.Upload(); err != nil {
			handleError(err)
			uploader.QuitDriver() // ✅ gracefully close Chrome
			continue
		}

		uploader.QuitDriver()

		videosUploaded++
		videosDayUploaded++
		scheduledTimeHour += scheduledIncrement

		fmt.Printf("✅ Uploaded video %d on %s at %s\n", videosUploaded, dateStr, scheduledTime)
		time.Sleep(2 * time.Second)
	}
}

	// handleError logs the error, prints stack trace, and tries to clean up Chrome.
func handleError(err error) {
	log.Printf("⚠️ An error occurred: %v\n", err)
	debug.PrintStack()

	// Fallback: kill Chrome if stuck
	if err := exec.Command("pkill", "-f", "chrome").Run(); err != nil {
		log.Printf("Failed to kill Chrome: %v", err)
	}
}



func demo(){
		type YouTubeSegmentDownloader struct {
		APIKey     string
		Client     *openai.Client
		OutputPath string
	}

	// Constructor with explicit API key
	// k := func NewYouTubeSegmentDownloader(apiKey string) *YouTubeSegmentDownloader {
	// 	client := openai.NewClient(apiKey)
	// 	return &YouTubeSegmentDownloader{APIKey: apiKey, Client: client}
	// }

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

}


func main() {

	utils.Init()

	lg := utils.Logger

	lg.Println("hello sts...")

	// download setup
	utils.SetupFFmpeg()
	utils.SetupModel()

	lg.Println("processing videos")

	stt.ProcessAllVideos(lg)

	text :="hello hi             bonjour "
	//voice := tts.Voice("en_us_001")
	outputFile := "output.mp3"

	err := tts.TTS(text, tts.Voice(models.UK_MALE_1), outputFile, false, lg)
	if err != nil {
		log.Fatalf("TTS error: %v", err)
	}

	lg.Println("TTS completed, saved to", outputFile)
}