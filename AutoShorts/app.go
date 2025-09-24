package autoshorts

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"time"

	"github.com/joho/godotenv"
)

var (
	openaiAPIKey string
	ytEmail      string
	ytPassword   string
	projectPath  string
)

func init() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ No .env file found, relying on system environment")
	}

	openaiAPIKey = os.Getenv("OPENAI_API_KEY")
	ytEmail = os.Getenv("YT_EMAIL")       // ✅ aligned with uploads.go
	ytPassword = os.Getenv("YT_PASSWORD") // ✅ aligned with uploads.go
	projectPath = os.Getenv("PROJECT_PATH")

	if openaiAPIKey == "" || ytEmail == "" || ytPassword == "" || projectPath == "" {
		log.Fatal("❌ Critical environment variables are missing. Please check your .env file.")
	}
}

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
