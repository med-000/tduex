package appconfig

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/med-000/tduscheexport/pkg/service"
)

type RuntimeConfig struct {
	Year   int
	Term   int
	Day    int
	Period int
	Mode   string
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
	if req.Year <= 0 {
		return service.GetCourseRequest{}, "", fmt.Errorf("year must be greater than 0")
	}
	if req.Term <= 0 {
		return service.GetCourseRequest{}, "", fmt.Errorf("term must be greater than 0")
	}
	if req.Day < 0 || req.Day > 7 {
		return service.GetCourseRequest{}, "", fmt.Errorf("day must be between 0 and 7")
	}
	if req.Period < 0 {
		return service.GetCourseRequest{}, "", fmt.Errorf("period must be 0 or greater")
	}

	return req, defaultOutputPath(cfg), nil
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

func defaultOutputPath(cfg RuntimeConfig) string {
	name := fmt.Sprintf("%s_%d_%d", cfg.Mode, cfg.Year, cfg.Term)
	if cfg.Day > 0 {
		name = fmt.Sprintf("%s_day%d", name, cfg.Day)
	}
	if cfg.Period > 0 {
		name = fmt.Sprintf("%s_period%d", name, cfg.Period)
	}
	return filepath.Join("out", name+".json")
}
