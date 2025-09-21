package stt

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
	"github.com/go-audio/wav"

	"sts/internal/models"
)

// ProcessAllVideos scans the Video folder and processes each video file
func ProcessAllVideos(logger *log.Logger) error {
	videoDir := "Video"
	audioDir := "audio"
	ttsDir := "stt"
	modelPath := "models/ggml-base.en.bin"

	// Ensure required folders exist
	for _, dir := range []string{videoDir, audioDir, ttsDir} {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create folder %s: %v", dir, err)
			}
			logger.Printf("Created folder: %s", dir)
		}
	}

	// Read video files
	files, err := os.ReadDir(videoDir)
	if err != nil {
		return fmt.Errorf("failed to read video folder: %v", err)
	}

	if len(files) == 0 {
		logger.Println("No video files found in Video/")
		return nil
	}

	// Process all video files
	for _, file := range files {
		if !file.IsDir() && (strings.HasSuffix(strings.ToLower(file.Name()), ".mp4") ||
			strings.HasSuffix(strings.ToLower(file.Name()), ".mov") ||
			strings.HasSuffix(strings.ToLower(file.Name()), ".mkv")) {

			videoPath := filepath.Join(videoDir, file.Name())
			if err := ProcessSingleVideo(videoPath, audioDir, ttsDir, modelPath, logger); err != nil {
				logger.Printf("Error processing %s: %v", file.Name(), err)
			}
		}
	}

	return nil
}

// ProcessSingleVideo handles one video: extract audio, transcribe, save JSON
func ProcessSingleVideo(videoPath, audioDir, ttsDir, modelPath string, logger *log.Logger) error {
	videoName := strings.TrimSuffix(filepath.Base(videoPath), filepath.Ext(videoPath))
	audioFile := filepath.Join(audioDir, videoName+".wav")
	jsonFile := filepath.Join(ttsDir, videoName+".json")

	// Skip if JSON already exists
	if _, err := os.Stat(jsonFile); err == nil {
		logger.Printf("Skipping %s (already processed)", videoName)
		return nil
	}

	logger.Printf("Processing video: %s", videoPath)

	// Step 1: Extract audio with ffmpeg
	cmd := exec.Command("ffmpeg", "-i", videoPath, "-ar", "16000", "-ac", "1", "-f", "wav", audioFile)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract audio: %v", err)
	}
	logger.Printf("Extracted audio: %s", audioFile)

	// Step 2: Load Whisper model
	model, err := whisper.New(modelPath)
	if err != nil {
		return fmt.Errorf("failed to load model: %v", err)
	}
	defer model.Close()

	// Step 3: Create a new context
	ctx, err := model.NewContext()
	if err != nil {
		return fmt.Errorf("failed to create whisper context: %v", err)
	}

	// Step 2: Read wav to []float32 (resampling already done by ffmpeg above)
	samples, sr, err := readWavToFloat32(audioFile)
	if err != nil {
		return fmt.Errorf("failed to read wav: %w", err)
	}
	if sr != 16000 {
		// defensive: if sample rate is not 16k, resample using ffmpeg and re-read
		logger.Printf("resampling audio from %d -> 16000", sr)
		tmp := audioFile + ".16k.wav"
		cmd := exec.Command("ffmpeg", "-y", "-i", audioFile, "-ar", "16000", "-ac", "1", tmp)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("ffmpeg resample failed: %w", err)
		}
		// replace audioFile and re-read
		audioFile = tmp
		samples, sr, err = readWavToFloat32(audioFile)
		if err != nil {
			return fmt.Errorf("failed to re-read resampled wav: %w", err)
		}
		if sr != 16000 {
			return fmt.Errorf("unexpected sample rate after resample: %d", sr)
		}
	}

	// Step 5: Run transcription
	if err := ctx.Process(samples, nil, nil, nil); err != nil {
		return fmt.Errorf("failed to process audio: %v", err)
	}

	// Step 6: Collect results
	var results []models.SegmentResult
	for {
		segment, err := ctx.NextSegment()
		if err != nil {
			break // no more segments
		}
		results = append(results, models.SegmentResult{
			Start: segment.Start,
			End:   segment.End,
			Text:  segment.Text,
		})
	}

	// Step 7: Save JSON output
	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	if err := os.WriteFile(jsonFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file: %v", err)
	}
	logger.Printf("Saved transcription to: %s", jsonFile)

	return nil
}

// readWavToFloat32 reads a wav file (PCM) and returns mono float32 samples and sample rate.
// If the file has multiple channels it averages them into mono.
// Uses go-audio/wav FullPCMBuffer to get PCM data. See package docs for FullPCMBuffer. :contentReference[oaicite:2]{index=2}
func readWavToFloat32(path string) ([]float32, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, fmt.Errorf("open wav: %w", err)
	}
	defer f.Close()

	dec := wav.NewDecoder(f)
	if !dec.IsValidFile() {
		return nil, 0, fmt.Errorf("invalid wav file: %s", path)
	}

	buf, err := dec.FullPCMBuffer()
	if err != nil {
		return nil, 0, fmt.Errorf("reading full pcm buffer: %w", err)
	}

	channels := buf.Format.NumChannels
	sr := buf.Format.SampleRate
	bitDepth := int(dec.SampleBitDepth())
	if bitDepth == 0 {
		// fallback to 16-bit if decoder can't tell us
		bitDepth = 16
	}
	// scale factor to map integer PCM -> [-1.0, +1.0)
	scale := float32((int(1) << (bitDepth - 1)))

	// buf.Data is interleaved samples (len = frames * channels)
	if channels <= 1 {
		samples := make([]float32, len(buf.Data))
		for i, v := range buf.Data {
			samples[i] = float32(v) / scale
		}
		return samples, sr, nil
	}

	frames := len(buf.Data) / channels
	samples := make([]float32, frames)
	for i := 0; i < frames; i++ {
		var sum int
		for ch := 0; ch < channels; ch++ {
			sum += buf.Data[i*channels+ch]
		}
		avg := float32(sum) / float32(channels)
		samples[i] = avg / scale
	}
	return samples, sr, nil
}
