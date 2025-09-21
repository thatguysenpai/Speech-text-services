package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
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
