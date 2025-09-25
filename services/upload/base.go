package upload

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/joho/godotenv"
)

// YouTubeVideoUploader encapsulates upload configuration
type YouTubeVideoUploader struct {
	Email    string
	Password string
	FilePath string
	Date     string // format: YYYY-MM-DD
	Time     string // format: HH:MM
	Ctx      context.Context
	Cancel   context.CancelFunc
}

// SetupDriver initializes chromedp with a Chrome profile
func (u *YouTubeVideoUploader) SetupDriver() {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserDataDir(os.Getenv("CHROME_PROFILE_PATH")),
		// chromedp.Flag("headless", true), // Uncomment for headless mode
	)

	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))

	u.Ctx = ctx
	u.Cancel = cancel
}

// Upload runs the YouTube upload flow
func (u *YouTubeVideoUploader) Upload() error {
	tasks := chromedp.Tasks{
		chromedp.Navigate("https://studio.youtube.com"),
		chromedp.Sleep(10 * time.Second),

		// Click Create button
		chromedp.Click(`#create-icon`, chromedp.NodeVisible),
		chromedp.Sleep(2 * time.Second),

		// Click Upload Videos
		chromedp.Click(`#text-item-0`, chromedp.NodeVisible),
		chromedp.Sleep(2 * time.Second),

		// Upload file
		chromedp.SetUploadFiles(`input[name="Filedata"]`, []string{u.FilePath}),
		chromedp.Sleep(20 * time.Second), // allow upload start

		// Click Next buttons
		chromedp.Click(`#next-button`, chromedp.NodeVisible),
		chromedp.Sleep(2 * time.Second),
		chromedp.Click(`#next-button`, chromedp.NodeVisible),
		chromedp.Sleep(2 * time.Second),
		chromedp.Click(`#next-button`, chromedp.NodeVisible),
		chromedp.Sleep(2 * time.Second),

		// Choose "Schedule"
		chromedp.Click(`div[aria-label="Schedule"]`, chromedp.NodeVisible),
		chromedp.Sleep(2 * time.Second),

		// Input date + time
		chromedp.SendKeys(`//input[@type="date"]`, u.Date),
		chromedp.Sleep(1 * time.Second),
		chromedp.SendKeys(`//input[@type="time"]`, u.Time),
		chromedp.Sleep(1 * time.Second),

		// Confirm scheduling
		chromedp.Click(`#done-button`, chromedp.NodeVisible),
	}

	log.Printf("[INFO] Uploading %s scheduled for %s %s", u.FilePath, u.Date, u.Time)
	return chromedp.Run(u.Ctx, tasks)
}

// QuitDriver closes the browser and kills leftover processes
func (u *YouTubeVideoUploader) QuitDriver() {
	if u.Cancel != nil {
		u.Cancel()
	}
	// Kill chrome if stuck
	exec.Command("pkill", "chrome").Run()
}

// LoadUploaderFromEnv is a helper to initialize uploader from .env
func LoadUploaderFromEnv(filePath, date, timeStr string) YouTubeVideoUploader {
	if err := godotenv.Load(); err != nil {
		log.Println("[WARN] No .env file found, relying on system environment")
	}

	return YouTubeVideoUploader{
		Email:    os.Getenv("YT_EMAIL"),
		Password: os.Getenv("YT_PASSWORD"),
		FilePath: filePath,
		Date:     date,
		Time:     timeStr,
	}
}
