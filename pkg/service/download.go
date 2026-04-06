package service

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gocolly/colly"
)

func DownloadFile(c *colly.Collector, url string, saveDir string) error {
	var fileData []byte
	var contentType string
	var filename string

	collector := c.Clone()

	collector.OnResponse(func(r *colly.Response) {
		fileData = r.Body
		contentType = r.Headers.Get("Content-Type")

		cd := r.Headers.Get("Content-Disposition")
		filename = extractFilename(cd)
	})

	if err := collector.Visit(url); err != nil {
		return err
	}
	collector.Wait()

	if len(fileData) == 0 {
		return fmt.Errorf("empty file response")
	}

	if !strings.Contains(contentType, "application") {
		return fmt.Errorf("not a file: %s", contentType)
	}

	if filename == "" {
		filename = "file.pdf"
	}

	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return err
	}

	path := filepath.Join(saveDir, filename)

	return os.WriteFile(path, fileData, 0644)
}

func extractFilename(cd string) string {
	re := regexp.MustCompile(`filename="?(.+?)"?$`)
	m := re.FindStringSubmatch(cd)
	if len(m) > 1 {
		return m[1]
	}
	return ""
}
