package db

import (
	"path/filepath"
	"testing"
	"time"
)

func TestStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "riptide.db")
	s, err := OpenPath(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	if err := s.SetSetting("theme", "ocean"); err != nil {
		t.Fatal(err)
	}
	if got := s.GetSetting("theme", "default"); got != "ocean" {
		t.Fatalf("theme=%q", got)
	}

	id, err := s.SaveTestRun(TestRun{
		Name:         "Evening",
		Kind:         "speed",
		DownloadMbps: 200,
		UploadMbps:   40,
		PingMs:       12,
		CreatedAt:    time.Now(),
	})
	if err != nil || id == 0 {
		t.Fatalf("save: id=%d err=%v", id, err)
	}

	runs, err := s.LatestRuns(10)
	if err != nil || len(runs) != 1 {
		t.Fatalf("latest: n=%d err=%v", len(runs), err)
	}
	if runs[0].Name != "Evening" {
		t.Fatalf("name=%q", runs[0].Name)
	}

	if err := s.Reset(false); err != nil {
		t.Fatal(err)
	}
	n, _ := s.CountRuns()
	if n != 0 {
		t.Fatalf("count after reset=%d", n)
	}
	if got := s.GetSetting("theme", "default"); got != "ocean" {
		t.Fatalf("theme wiped: %q", got)
	}
}
