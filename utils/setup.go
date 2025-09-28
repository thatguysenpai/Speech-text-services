package utils

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// SetupFFmpeg checks if ffmpeg is installed and installs it if missing.
func SetupFFmpeg() {
	// Check if ffmpeg exists
	_, err := exec.LookPath("ffmpeg")
	if err == nil {
		fmt.Println("ffmpeg is already installed ✅")
		return
	}

	fmt.Println("ffmpeg not found, installing...")

	switch runtime.GOOS {
	case "linux":
		installLinux()
	case "darwin":
		installMac()
	case "windows":
		installWindows()
	default:
		log.Fatalf("Unsupported OS: %s. Please install ffmpeg manually.", runtime.GOOS)
	}
}

func installLinux() {
	cmd := exec.Command("bash", "-c", "sudo apt-get update && sudo apt-get install -y ffmpeg")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to install ffmpeg on Linux: %v", err)
	}
}

func installMac() {
	cmd := exec.Command("brew", "install", "ffmpeg")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to install ffmpeg on macOS: %v", err)
	}
}

func installWindows() {
	fmt.Println("Please install ffmpeg manually on Windows via https://ffmpeg.org/download.html or Chocolatey:")
	cmd := exec.Command("choco", "install", "ffmpeg", "-y")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to install ffmpeg on Windows: %v", err)
	}
}

// SetupModel checks if the Whisper model exists, if not downloads it
func SetupModel() error {

	modelPath := "models/ggml-base.en.bin"
	if _, err := os.Stat(modelPath); err == nil {
		fmt.Println("Model already exists ✅:", modelPath)
		return nil
	}

	fmt.Println("Model not found, downloading...")

	// Create directory if not exists
	if err := os.MkdirAll(filepath.Dir(modelPath), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create model directory: %v", err)
	}

	url := "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.en.bin"
	out, err := os.Create(modelPath)
	if err != nil {
		return fmt.Errorf("failed to create model file: %v", err)
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download model: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status downloading model: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save model: %v", err)
	}

	fmt.Println("Model downloaded successfully ✅:", modelPath)

	return nil
}

func SetupWhisper(logger *log.Logger) error {

	repoDir := "whisper.cpp"
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		logger.Println("whisper.cpp not found, cloning...")
		cmd := exec.Command("git", "clone", "https://github.com/ggerganov/whisper.cpp.git", repoDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to clone whisper.cpp: %v", err)
		}
	} else {
		logger.Println("whisper.cpp already exists")
	}

	makeCmd := exec.Command("make")
	makeCmd.Dir = repoDir
	makeCmd.Stdout = os.Stdout
	makeCmd.Stderr = os.Stderr
	if err := makeCmd.Run(); err != nil {
		return fmt.Errorf("failed to build whisper.cpp: %v", err)
	}
	logger.Println("whisper.cpp built successfully")
	absRepo, err := filepath.Abs(repoDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for whisper.cpp: %v", err)
	}

	os.Setenv("CGO_CFLAGS", "-I"+absRepo)
	os.Setenv("CGO_LDFLAGS", "-L"+absRepo+" -lwhisper")
	logger.Printf("Set CGO_CFLAGS=%s", os.Getenv("CGO_CFLAGS"))
	logger.Printf("Set CGO_LDFLAGS=%s", os.Getenv("CGO_LDFLAGS"))

	return nil
}

func SetupFFprobe(logger *log.Logger) {
	// Check if ffprobe exists
	_, err := exec.LookPath("ffprobe")
	if err == nil {
		logger.Println("ffprobe is already installed ✅")
		return
	}

	logger.Println("ffprobe not found, installing...")

	switch runtime.GOOS {
	case "linux":
		installFFprobeLinux()
	case "darwin":
		installFFprobeMac()
	case "windows":
		installFFprobeWindows()
	default:
		log.Fatalf("Unsupported OS: %s. Please install ffprobe manually.", runtime.GOOS)
	}
}

func installFFprobeLinux() {
	cmd := exec.Command("bash", "-c", "sudo apt-get update && sudo apt-get install -y ffmpeg")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to install ffprobe on Linux: %v", err)
	}
}

func installFFprobeMac() {
	cmd := exec.Command("brew", "install", "ffmpeg")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to install ffprobe on macOS: %v", err)
	}
}

func installFFprobeWindows() {
	fmt.Println("Please install ffprobe manually on Windows via https://ffmpeg.org/download.html or Chocolatey:")
	cmd := exec.Command("choco", "install", "ffmpeg", "-y")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to install ffprobe on Windows: %v", err)
	}
}
