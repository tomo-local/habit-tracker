package sessionlog

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeSession(t *testing.T, dir, name string, lines []string) {
	t.Helper()
	content := ""
	for _, l := range lines {
		content += l + "\n"
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestBlocks(t *testing.T) {
	dir := t.TempDir()

	// Block 1: 10:00 - 10:05, one prompt (duplicated line should be deduped).
	// Block 2: 11:00 (gap > GapThreshold from block 1), one prompt.
	writeSession(t, dir, "session.jsonl", []string{
		`{"type":"user","timestamp":"2026-01-01T10:00:00Z","uuid":"u1"}`,
		`{"type":"last-prompt","lastPrompt":"do the thing","leafUuid":"u1"}`,
		`{"type":"last-prompt","lastPrompt":"do the thing","leafUuid":"u1"}`,
		`{"type":"assistant","timestamp":"2026-01-01T10:05:00Z","uuid":"u2"}`,
		`{"type":"user","timestamp":"2026-01-01T11:00:00Z","uuid":"u3"}`,
		`{"type":"last-prompt","lastPrompt":"do another thing","leafUuid":"u3"}`,
	})

	blocks, latest, err := Blocks(dir, time.Time{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(blocks) != 2 {
		t.Fatalf("got %d blocks, want 2", len(blocks))
	}

	b1 := blocks[0]
	if !b1.Start.Equal(time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)) {
		t.Errorf("block1 start = %v", b1.Start)
	}
	if !b1.End.Equal(time.Date(2026, 1, 1, 10, 5, 0, 0, time.UTC)) {
		t.Errorf("block1 end = %v", b1.End)
	}
	if len(b1.Notes) != 1 || b1.Notes[0] != "do the thing" {
		t.Errorf("block1 notes = %v, want deduped [\"do the thing\"]", b1.Notes)
	}
	if got := b1.Description(); got != "- do the thing" {
		t.Errorf("block1 description = %q", got)
	}

	b2 := blocks[1]
	if len(b2.Notes) != 1 || b2.Notes[0] != "do another thing" {
		t.Errorf("block2 notes = %v", b2.Notes)
	}

	wantLatest := time.Date(2026, 1, 1, 11, 0, 0, 0, time.UTC)
	if !latest.Equal(wantLatest) {
		t.Errorf("latest = %v, want %v", latest, wantLatest)
	}

	// Since watermark excludes block1 (started before/at since) and keeps block2.
	since := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)
	filtered, _, err := Blocks(dir, since)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(filtered) != 1 {
		t.Fatalf("got %d blocks after watermark, want 1", len(filtered))
	}
	if !filtered[0].Start.Equal(b2.Start) {
		t.Errorf("filtered block start = %v, want %v", filtered[0].Start, b2.Start)
	}
}

func TestProjectDir(t *testing.T) {
	dir, err := ProjectDir("/Users/x/proj")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filepath.Base(dir) != "-Users-x-proj" {
		t.Errorf("got %q, want suffix -Users-x-proj", dir)
	}
}
