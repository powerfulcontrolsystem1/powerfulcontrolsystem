package logstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/you/pos-backend/vpssecurity/reports"
)

type Store struct {
	root string
	mu   sync.Mutex
}

func NewStore(root string) *Store {
	return &Store{root: filepath.Clean(root)}
}

func (s *Store) Root() string {
	return s.root
}

func (s *Store) Ensure() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return os.MkdirAll(filepath.Join(s.root, "runs"), 0o700)
}

func (s *Store) Save(report *reports.ScanReport, generated map[string][]byte, rawArtifacts map[string][]byte) error {
	if report == nil {
		return errors.New("report is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	runDir := filepath.Join(s.root, "runs", report.ScanID)
	reportsDir := filepath.Join(runDir, "reports")
	rawDir := filepath.Join(runDir, "raw")
	if err := os.MkdirAll(reportsDir, 0o700); err != nil {
		return err
	}
	if err := os.MkdirAll(rawDir, 0o700); err != nil {
		return err
	}
	for format, content := range generated {
		fileName := reportFileName(format)
		if err := os.WriteFile(filepath.Join(reportsDir, fileName), content, 0o600); err != nil {
			return err
		}
	}
	for name, content := range rawArtifacts {
		target := filepath.Join(runDir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			return err
		}
		if err := os.WriteFile(target, content, 0o600); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) Load(scanID string) (*reports.ScanReport, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadUnlocked(scanID)
}

func (s *Store) loadUnlocked(scanID string) (*reports.ScanReport, error) {
	path := filepath.Join(s.root, "runs", scanID, "reports", reportFileName("json"))
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var report reports.ScanReport
	if err := jsonUnmarshal(raw, &report); err != nil {
		return nil, err
	}
	return &report, nil
}

func (s *Store) List(limit int) ([]reports.HistoryEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	runsDir := filepath.Join(s.root, "runs")
	entries, err := os.ReadDir(runsDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	history := make([]reports.HistoryEntry, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		report, err := s.loadUnlocked(entry.Name())
		if err != nil || report == nil {
			continue
		}
		history = append(history, reports.HistoryFromReport(report))
	}
	sort.Slice(history, func(i, j int) bool {
		return history[i].GeneratedAt > history[j].GeneratedAt
	})
	if limit > 0 && len(history) > limit {
		history = history[:limit]
	}
	return history, nil
}

func (s *Store) LoadLatest() (*reports.ScanReport, error) {
	history, err := s.List(1)
	if err != nil || len(history) == 0 {
		return nil, err
	}
	return s.Load(history[0].ScanID)
}

func (s *Store) LoadPreviousBefore(scanID string) (*reports.ScanReport, error) {
	history, err := s.List(0)
	if err != nil {
		return nil, err
	}
	seen := false
	for _, item := range history {
		if seen {
			return s.Load(item.ScanID)
		}
		if item.ScanID == scanID {
			seen = true
		}
	}
	return nil, nil
}

func (s *Store) LoadArtifact(scanID, format string) ([]byte, string, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fileName := reportFileName(format)
	if fileName == "" {
		return nil, "", "", fmt.Errorf("formato no soportado: %s", format)
	}
	path := filepath.Join(s.root, "runs", scanID, "reports", fileName)
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, "", "", err
	}
	return raw, fileName, contentType(format), nil
}

func reportFileName(format string) string {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		return "report.json"
	case "txt":
		return "report.txt"
	case "html":
		return "report.html"
	case "csv":
		return "report.csv"
	case "pdf":
		return "report.pdf"
	case "xls", "excel":
		return "report.xls"
	default:
		return ""
	}
}

func contentType(format string) string {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		return "application/json"
	case "txt":
		return "text/plain; charset=utf-8"
	case "html":
		return "text/html; charset=utf-8"
	case "csv":
		return "text/csv; charset=utf-8"
	case "pdf":
		return "application/pdf"
	case "xls", "excel":
		return "application/vnd.ms-excel"
	default:
		return "application/octet-stream"
	}
}

func jsonUnmarshal(data []byte, target interface{}) error {
	return json.Unmarshal(data, target)
}
