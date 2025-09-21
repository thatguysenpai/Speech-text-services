package models

import "time"

// SegmentResult represents transcription output with timestamps
type SegmentResult struct {
	Start time.Duration `json:"start"`
	End   time.Duration `json:"end"`
	Text  string        `json:"text"`
}
