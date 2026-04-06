package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	maxLogSizeBytes     = 5 * 1024 * 1024
	maxDailyGenerations = 7
	maxSizeGenerations  = 5
)

var (
	writerRegistryMu sync.Mutex
	writerRegistry   = map[string]*rotatingFileWriter{}
)

type rotatingFileWriter struct {
	path        string
	mu          sync.Mutex
	file        *os.File
	currentDate string
	currentSize int64
}

type backupInfo struct {
	path  string
	date  string
	index int
}

func getRotatingWriter(path string) (io.Writer, error) {
	writerRegistryMu.Lock()
	defer writerRegistryMu.Unlock()

	if writer, ok := writerRegistry[path]; ok {
		return writer, nil
	}

	writer, err := newRotatingFileWriter(path)
	if err != nil {
		return nil, err
	}

	writerRegistry[path] = writer
	return writer, nil
}

func newRotatingFileWriter(path string) (*rotatingFileWriter, error) {
	writer := &rotatingFileWriter{
		path:        path,
		currentDate: time.Now().Format("2006-01-02"),
	}

	if err := writer.openCurrentFile(); err != nil {
		return nil, err
	}

	return writer, nil
}

func (w *rotatingFileWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	today := time.Now().Format("2006-01-02")
	needsDailyRotate := today != w.currentDate
	needsSizeRotate := w.currentSize+int64(len(p)) > maxLogSizeBytes

	if needsDailyRotate || needsSizeRotate {
		if err := w.rotate(today); err != nil {
			return 0, err
		}
	}

	n, err := w.file.Write(p)
	w.currentSize += int64(n)
	return n, err
}

func (w *rotatingFileWriter) rotate(nextDate string) error {
	if w.file != nil {
		if err := w.file.Close(); err != nil {
			return err
		}
	}

	if err := w.archiveCurrentFile(); err != nil {
		return err
	}

	w.currentDate = nextDate
	if err := w.openCurrentFile(); err != nil {
		return err
	}

	return w.pruneBackups()
}

func (w *rotatingFileWriter) archiveCurrentFile() error {
	info, err := os.Stat(w.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if info.Size() == 0 {
		return os.Remove(w.path)
	}

	date := w.currentDate
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	index, err := w.nextBackupIndex(date)
	if err != nil {
		return err
	}

	backupPath := fmt.Sprintf("%s.%s.%02d", w.path, date, index)
	return os.Rename(w.path, backupPath)
}

func (w *rotatingFileWriter) openCurrentFile() error {
	if err := os.MkdirAll(filepath.Dir(w.path), 0o755); err != nil {
		return err
	}

	file, err := os.OpenFile(w.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return err
	}

	w.file = file
	w.currentSize = info.Size()
	return nil
}

func (w *rotatingFileWriter) nextBackupIndex(date string) (int, error) {
	backups, err := w.listBackups()
	if err != nil {
		return 0, err
	}

	maxIndex := 0
	for _, backup := range backups {
		if backup.date == date && backup.index > maxIndex {
			maxIndex = backup.index
		}
	}

	return maxIndex + 1, nil
}

func (w *rotatingFileWriter) pruneBackups() error {
	backups, err := w.listBackups()
	if err != nil {
		return err
	}

	dates := distinctDates(backups)
	if len(dates) > maxDailyGenerations {
		for _, expiredDate := range dates[:len(dates)-maxDailyGenerations] {
			for _, backup := range backups {
				if backup.date == expiredDate {
					if err := os.Remove(backup.path); err != nil && !os.IsNotExist(err) {
						return err
					}
				}
			}
		}
	}

	backups, err = w.listBackups()
	if err != nil {
		return err
	}

	grouped := make(map[string][]backupInfo)
	for _, backup := range backups {
		grouped[backup.date] = append(grouped[backup.date], backup)
	}

	for _, group := range grouped {
		sort.Slice(group, func(i, j int) bool {
			return group[i].index < group[j].index
		})

		if len(group) <= maxSizeGenerations {
			continue
		}

		for _, expired := range group[:len(group)-maxSizeGenerations] {
			if err := os.Remove(expired.path); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
	}

	return nil
}

func (w *rotatingFileWriter) listBackups() ([]backupInfo, error) {
	matches, err := filepath.Glob(w.path + ".*")
	if err != nil {
		return nil, err
	}

	backups := make([]backupInfo, 0, len(matches))
	for _, match := range matches {
		date, index, ok := parseBackupName(w.path, match)
		if !ok {
			continue
		}

		backups = append(backups, backupInfo{
			path:  match,
			date:  date,
			index: index,
		})
	}

	return backups, nil
}

func parseBackupName(basePath, backupPath string) (string, int, bool) {
	prefix := basePath + "."
	if !strings.HasPrefix(backupPath, prefix) {
		return "", 0, false
	}

	rest := strings.TrimPrefix(backupPath, prefix)
	parts := strings.Split(rest, ".")
	if len(parts) != 4 {
		return "", 0, false
	}

	date := strings.Join(parts[:3], "-")
	index, err := strconv.Atoi(parts[3])
	if err != nil {
		return "", 0, false
	}

	return date, index, true
}

func distinctDates(backups []backupInfo) []string {
	set := make(map[string]struct{})
	for _, backup := range backups {
		set[backup.date] = struct{}{}
	}

	dates := make([]string, 0, len(set))
	for date := range set {
		dates = append(dates, date)
	}

	sort.Strings(dates)
	return dates
}
