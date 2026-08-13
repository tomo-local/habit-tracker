// Package sessionlog reads Claude Code's local session transcripts
// (~/.claude/projects/<slug>/*.jsonl) and groups them into contiguous work
// blocks, so habit-tracker can back-fill calendar events for AI-assisted
// work sessions.
package sessionlog

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// GapThreshold is the idle time after which a new work block starts.
const GapThreshold = 15 * time.Minute

// Block is a contiguous burst of session activity.
type Block struct {
	Start time.Time
	End   time.Time
	Notes []string
}

// ProjectDir returns the Claude Code session log directory for the given
// working directory, e.g. "/Users/x/proj" -> "~/.claude/projects/-Users-x-proj".
func ProjectDir(cwd string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	slug := strings.ReplaceAll(cwd, "/", "-")
	return filepath.Join(home, ".claude", "projects", slug), nil
}

// Blocks reads every session log under dir and returns the work blocks that
// started after since, along with the latest timestamp seen across all
// logs (to be persisted as the new watermark).
func Blocks(dir string, since time.Time) ([]Block, time.Time, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.jsonl"))
	if err != nil {
		return nil, since, err
	}

	var all []Block
	latest := since
	for _, file := range files {
		blocks, fileLatest, err := parseFile(file)
		if err != nil {
			return nil, since, err
		}
		all = append(all, blocks...)
		if fileLatest.After(latest) {
			latest = fileLatest
		}
	}

	sort.Slice(all, func(i, j int) bool { return all[i].Start.Before(all[j].Start) })

	var result []Block
	for _, b := range all {
		if b.Start.After(since) {
			result = append(result, b)
		}
	}
	return result, latest, nil
}

type rawEntry struct {
	Type       string `json:"type"`
	Timestamp  string `json:"timestamp"`
	UUID       string `json:"uuid"`
	LastPrompt string `json:"lastPrompt"`
	LeafUUID   string `json:"leafUuid"`
}

func parseFile(path string) ([]Block, time.Time, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, time.Time{}, err
	}
	defer f.Close()

	byUUID := make(map[string]time.Time)
	var lastPrompts []rawEntry
	var timestamps []time.Time
	var latest time.Time

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	for scanner.Scan() {
		var e rawEntry
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			continue
		}
		if e.Timestamp != "" {
			if ts, err := time.Parse(time.RFC3339, e.Timestamp); err == nil {
				timestamps = append(timestamps, ts)
				if e.UUID != "" {
					byUUID[e.UUID] = ts
				}
				if ts.After(latest) {
					latest = ts
				}
			}
		}
		if e.Type == "last-prompt" {
			lastPrompts = append(lastPrompts, e)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, latest, err
	}

	sort.Slice(timestamps, func(i, j int) bool { return timestamps[i].Before(timestamps[j]) })
	blocks := clusterTimestamps(timestamps)

	for _, lp := range lastPrompts {
		ts, ok := byUUID[lp.LeafUUID]
		if !ok {
			continue
		}
		for i := range blocks {
			if !ts.Before(blocks[i].Start) && !ts.After(blocks[i].End) {
				note := collapseWhitespace(lp.LastPrompt)
				if !contains(blocks[i].Notes, note) {
					blocks[i].Notes = append(blocks[i].Notes, note)
				}
				break
			}
		}
	}

	return blocks, latest, nil
}

func contains(notes []string, note string) bool {
	for _, n := range notes {
		if n == note {
			return true
		}
	}
	return false
}

func clusterTimestamps(ts []time.Time) []Block {
	var blocks []Block
	for _, t := range ts {
		if len(blocks) > 0 && t.Sub(blocks[len(blocks)-1].End) <= GapThreshold {
			blocks[len(blocks)-1].End = t
			continue
		}
		blocks = append(blocks, Block{Start: t, End: t})
	}
	return blocks
}

func collapseWhitespace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// Description renders a block's notes as a bullet list for a calendar
// event's description field.
func (b Block) Description() string {
	if len(b.Notes) == 0 {
		return ""
	}
	lines := make([]string, len(b.Notes))
	for i, n := range b.Notes {
		lines[i] = "- " + n
	}
	return strings.Join(lines, "\n")
}
