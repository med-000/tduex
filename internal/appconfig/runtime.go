package appconfig

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/med-000/tduscheexport/internal/service"
)

type RuntimeConfig struct {
	Year      int
	Term      int
	Day       int
	Period    int
	Mode      string
	Formats   []string
	UseDialog bool
}

func BuildRequest(cfg RuntimeConfig) (service.GetCourseRequest, string, error) {
	req := service.GetCourseRequest{
		UserID:   os.Getenv("USER_ID"),
		Password: os.Getenv("PASSWORD"),
		Year:     cfg.Year,
		Term:     cfg.Term,
		Day:      cfg.Day,
		Period:   cfg.Period,
	}

	if req.UserID == "" {
		return service.GetCourseRequest{}, "", fmt.Errorf("missing required env: USER_ID")
	}
	if req.Password == "" {
		return service.GetCourseRequest{}, "", fmt.Errorf("missing required env: PASSWORD")
	}
	if err := ValidateRuntimeConfig(cfg); err != nil {
		return service.GetCourseRequest{}, "", err
	}

	return req, BuildOutputPath(cfg, ".json"), nil
}

func ValidateRuntimeConfig(cfg RuntimeConfig) error {
	if err := ValidateYear(cfg.Year); err != nil {
		return err
	}
	if err := ValidateTerm(cfg.Term); err != nil {
		return err
	}
	if err := ValidateDay(cfg.Day); err != nil {
		return err
	}
	if err := ValidatePeriod(cfg.Period); err != nil {
		return err
	}
	return nil
}

func ValidateYear(year int) error {
	if year < 2000 || year > 2100 {
		return fmt.Errorf("year must be between 2000 and 2100")
	}
	return nil
}

func ValidateTerm(term int) error {
	if term < 1 || term > 2 {
		return fmt.Errorf("term must be between 1 and 2")
	}
	return nil
}

func ValidateDay(day int) error {
	if day < 0 || day > 7 {
		return fmt.Errorf("day must be between 0 and 7")
	}
	return nil
}

func ValidatePeriod(period int) error {
	if period < 0 || period > 7 {
		return fmt.Errorf("period must be between 0 and 7")
	}
	return nil
}

func LoadDotEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}

		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)

		if _, exists := os.LookupEnv(key); exists {
			continue
		}

		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}

	return scanner.Err()
}

func LoadOptionalEnv(path string) error {
	if err := LoadDotEnv(path); err != nil {
		if errorsIsNotExist(err) {
			return nil
		}
		return err
	}
	return nil
}

func PersistCredentials(path string, userID string, password string) error {
	if err := ensureParentDir(path); err != nil {
		return err
	}

	lines, err := readEnvLines(path)
	if err != nil && !errorsIsNotExist(err) {
		return err
	}

	lines = upsertEnvLine(lines, "USER_ID", userID)
	lines = upsertEnvLine(lines, "PASSWORD", password)

	data := strings.Join(lines, "\n")
	if data != "" {
		data += "\n"
	}

	return os.WriteFile(path, []byte(data), 0600)
}

func BuildOutputPath(cfg RuntimeConfig, ext string) string {
	name := fmt.Sprintf("%s_%d_%d", cfg.Mode, cfg.Year, cfg.Term)
	if cfg.Day > 0 {
		name = fmt.Sprintf("%s_day%d", name, cfg.Day)
	}
	if cfg.Period > 0 {
		name = fmt.Sprintf("%s_period%d", name, cfg.Period)
	}
	return filepath.Join("out", name+ext)
}

func ensureParentDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." {
		return nil
	}
	return os.MkdirAll(dir, 0755)
}

func errorsIsNotExist(err error) bool {
	return err != nil && (os.IsNotExist(err) || isPathErrorNotExist(err))
}

func isPathErrorNotExist(err error) bool {
	var pathErr *fs.PathError
	return err != nil && errors.As(err, &pathErr) && os.IsNotExist(pathErr)
}

func readEnvLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	text = strings.TrimRight(text, "\n")
	if text == "" {
		return []string{}, nil
	}
	return strings.Split(text, "\n"), nil
}

func upsertEnvLine(lines []string, key string, value string) []string {
	prefix := key + "="
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, prefix) {
			lines[i] = prefix + value
			return lines
		}
	}
	return append(lines, prefix+value)
}
