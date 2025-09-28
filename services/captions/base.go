package captions

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

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

// // CreateSubtitleFile writes SRT subtitles by distributing text evenly over audio duration.
// func CreateSubtitleFile(segments []openai.TranscriptionSegment, subtitlePath, audioPath string) error {
// 	duration, err := GetAudioDuration(audioPath)
// 	if err != nil {
// 		return fmt.Errorf("failed to get audio duration: %w", err)
// 	}

// 	f, err := os.Create(subtitlePath)
// 	if err != nil {
// 		return err
// 	}
// 	defer f.Close()

// 	text := strings.TrimSpace(segments[0].Text)
// 	words := strings.Fields(text)

// 	// Adjustable: number of words per subtitle block
// 	chunkSize := 6
// 	numChunks := (len(words) + chunkSize - 1) / chunkSize
// 	chunkDuration := duration / float64(numChunks)

// 	index := 1
// 	for i := 0; i < len(words); i += chunkSize {
// 		startTime := float64((i / chunkSize)) * chunkDuration
// 		endTime := startTime + chunkDuration

// 		chunk := strings.Join(words[i:min(i+chunkSize, len(words))], " ")

// 		_, err := fmt.Fprintf(f, "%d\n%s --> %s\n%s\n\n",
// 			index,
// 			FormatTimestamp(startTime),
// 			FormatTimestamp(endTime),
// 			chunk,
// 		)
// 		if err != nil {
// 			return err
// 		}
// 		index++
// 	}

// 	return nil
// }

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


func ProcessVideo(apiKey, videoPath string) error {
	audioPath := "temp_audio.wav"
	subtitlePath := "temp_subtitles.srt"
	outputVideoPath := "output_video.mp4"
	fontPath := "/path/to/your/font.ttf" // adjust this

	defer func() {
		_ = os.Remove(audioPath)
		_ = os.Remove(subtitlePath)
	}()

	// 4. Burn subtitles
	if err := BurnSubtitles(videoPath, subtitlePath, outputVideoPath, fontPath); err != nil {
		return fmt.Errorf("failed to burn subtitles: %w", err)
	}

	fmt.Println("✅ Process completed. Subtitled video:", outputVideoPath)
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
