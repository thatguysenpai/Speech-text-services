package autoshorts

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

// ExtractAudio extracts mono 16kHz audio from the video using FFmpeg.
func ExtractAudio(videoPath, audioPath string) error {
	cmd := exec.Command("ffmpeg",
		"-i", videoPath,
		"-acodec", "pcm_s16le",
		"-ac", "1",
		"-ar", "16000",
		audioPath,
		"-y", // overwrite output
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// GenerateCaptions uses OpenAI Whisper API to transcribe audio.
func GenerateCaptions(apiKey, audioPath string) ([]openai.TranscriptionSegment, error) {
	client := openai.NewClient(apiKey)

	req := openai.AudioRequest{
		Model:    openai.Whisper1, // Whisper model
		FilePath: audioPath,
	}
	resp, err := client.CreateTranscription(context.Background(), req)
	if err != nil {
		return nil, err
	}

	// OpenAI API currently returns plain text without timestamps.
	segment := openai.TranscriptionSegment{
		Text:  resp.Text,
		Start: 0.0,
		End:   0.0,
	}
	return []openai.TranscriptionSegment{segment}, nil
}

// GetAudioDuration returns the length of the audio file in seconds using ffprobe.
func GetAudioDuration(audioPath string) (float64, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		audioPath,
	)
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	durationStr := strings.TrimSpace(string(out))
	return strconv.ParseFloat(durationStr, 64)
}

// CreateSubtitleFile writes SRT subtitles by distributing text evenly over audio duration.
func CreateSubtitleFile(segments []openai.TranscriptionSegment, subtitlePath, audioPath string) error {
	duration, err := GetAudioDuration(audioPath)
	if err != nil {
		return fmt.Errorf("failed to get audio duration: %w", err)
	}

	f, err := os.Create(subtitlePath)
	if err != nil {
		return err
	}
	defer f.Close()

	text := strings.TrimSpace(segments[0].Text)
	words := strings.Fields(text)

	// Adjustable: number of words per subtitle block
	chunkSize := 6
	numChunks := (len(words) + chunkSize - 1) / chunkSize
	chunkDuration := duration / float64(numChunks)

	index := 1
	for i := 0; i < len(words); i += chunkSize {
		startTime := float64((i / chunkSize)) * chunkDuration
		endTime := startTime + chunkDuration

		chunk := strings.Join(words[i:min(i+chunkSize, len(words))], " ")

		_, err := fmt.Fprintf(f, "%d\n%s --> %s\n%s\n\n",
			index,
			FormatTimestamp(startTime),
			FormatTimestamp(endTime),
			chunk,
		)
		if err != nil {
			return err
		}
		index++
	}

	return nil
}

// FormatTimestamp converts seconds to SRT time format.
func FormatTimestamp(seconds float64) string {
	hours := int(seconds) / 3600
	minutes := (int(seconds) % 3600) / 60
	secs := int(seconds) % 60
	ms := int((seconds - float64(int(seconds))) * 1000)

	return fmt.Sprintf("%02d:%02d:%02d,%03d", hours, minutes, secs, ms)
}

// BurnSubtitles embeds subtitles into a video using FFmpeg.
func BurnSubtitles(videoPath, subtitlePath, outputVideoPath, fontPath string) error {
	subtitleFilter := fmt.Sprintf("subtitles=%s:force_style='Alignment=10,FontName=TheBoldFont-Bold,FontSize=15.5,PrimaryColour=&H00ffffff,OutlineColour=&H00000000,BorderStyle=1,Outline=1.3,Shadow=0'", subtitlePath)

	cmd := exec.Command("ffmpeg",
		"-i", videoPath,
		"-vf", subtitleFilter,
		outputVideoPath,
		"-y",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Main pipeline
func ProcessVideo(apiKey, videoPath string) error {
	audioPath := "temp_audio.wav"
	subtitlePath := "temp_subtitles.srt"
	outputVideoPath := "output_video.mp4"
	fontPath := "/path/to/your/font.ttf" // adjust this

	defer func() {
		_ = os.Remove(audioPath)
		_ = os.Remove(subtitlePath)
	}()

	// 1. Extract audio
	if err := ExtractAudio(videoPath, audioPath); err != nil {
		return fmt.Errorf("failed to extract audio: %w", err)
	}

	// 2. Transcribe audio
	segments, err := GenerateCaptions(apiKey, audioPath)
	if err != nil {
		return fmt.Errorf("failed to transcribe: %w", err)
	}

	// 3. Create subtitles
	if err := CreateSubtitleFile(segments, subtitlePath, audioPath); err != nil {
		return fmt.Errorf("failed to create subtitles: %w", err)
	}

	// 4. Burn subtitles
	if err := BurnSubtitles(videoPath, subtitlePath, outputVideoPath, fontPath); err != nil {
		return fmt.Errorf("failed to burn subtitles: %w", err)
	}

	fmt.Println("✅ Process completed. Subtitled video:", outputVideoPath)
	return nil
}

// Helper: min function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
