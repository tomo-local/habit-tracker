package config

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func withTempDir(t *testing.T) {
	t.Helper()
	t.Setenv(habitConfigDir, t.TempDir())
}

func Test_New_FileNotExist(t *testing.T) {
	withTempDir(t)

	cfg := New()
	if err := cfg.Read(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.CalendarID != "" || len(cfg.Habits) != 0 {
		t.Errorf("expected empty config, got %+v", cfg)
	}
}

func Test_Write_RoundTrip(t *testing.T) {
	withTempDir(t)

	original := &Config{
		CalendarID:   "cal123",
		CalendarName: "habit",
		Habits:       []string{"workout", "reading"},
	}
	if err := original.Write(); err != nil {
		t.Fatalf("Write: %v", err)
	}

	got := New()
	if err := got.Read(); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if got.CalendarID != original.CalendarID {
		t.Errorf("CalendarID: got %q, want %q", got.CalendarID, original.CalendarID)
	}
	if got.CalendarName != original.CalendarName {
		t.Errorf("CalendarName: got %q, want %q", got.CalendarName, original.CalendarName)
	}
	if len(got.Habits) != len(original.Habits) {
		t.Errorf("Habits len: got %d, want %d", len(got.Habits), len(original.Habits))
	}
}

func Test_Read_MalformedJSON(t *testing.T) {
	withTempDir(t)

	if err := os.MkdirAll(ConfigDir(), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ConfigDir(), "config.json"), []byte("{invalid"), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := &Config{}
	if err := cfg.Read(); err == nil {
		t.Error("expected error for malformed JSON, got nil")
	}
}

func Test_Write_CreatesDir(t *testing.T) {
	t.Setenv(habitConfigDir, filepath.Join(t.TempDir(), "nested", "dir"))

	cfg := &Config{Habits: []string{"workout"}}
	if err := cfg.Write(); err != nil {
		t.Fatalf("Write should create directory: %v", err)
	}
	if _, err := os.Stat(ConfigDir()); err != nil {
		t.Errorf("directory not created: %v", err)
	}
}

func Test_Write_FilePermissions(t *testing.T) {
	withTempDir(t)

	cfg := &Config{Habits: []string{"workout"}}
	if err := cfg.Write(); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(filepath.Join(ConfigDir(), "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("expected permission 0600, got %o", perm)
	}
}

func Test_ConcurrentReadWrite(t *testing.T) {
	withTempDir(t)

	cfg := &Config{CalendarID: "initial", Habits: []string{"workout"}}
	if err := cfg.Write(); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for range 10 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			_ = cfg.Read()
		}()
		go func() {
			defer wg.Done()
			_ = cfg.Write()
		}()
	}
	wg.Wait()
}
