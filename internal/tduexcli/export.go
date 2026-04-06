package tduexcli

import (
	"fmt"
	"slices"
	"strings"

	"github.com/med-000/tduscheexport/internal/appconfig"
	"github.com/med-000/tduscheexport/internal/service"
)

func exportClassesByFormat(svc *service.Service, result *service.ExportCourse, format string, path string) error {
	switch format {
	case "json":
		return wrapExportError("class list json", svc.ExportCourseJSON(result, path))
	case "csv":
		return wrapExportError("class list csv", svc.ExportCourseCSV(result, path))
	case "xlsx":
		return wrapExportError("class list xlsx", svc.ExportCourseXLSX(result, path))
	default:
		return fmt.Errorf("unsupported format for classes: %s", format)
	}
}

func exportFullByFormat(svc *service.Service, result *service.FullExportCourse, format string, path string) error {
	switch format {
	case "json":
		return wrapExportError("full json", svc.ExportFullCourseJSON(result, path))
	case "csv":
		return wrapExportError("full csv", svc.ExportFullCourseCSV(result, path))
	case "xlsx":
		return wrapExportError("full xlsx", svc.ExportFullCourseXLSX(result, path))
	case "ics":
		return wrapExportError("full ics", svc.ExportFullCourseICS(result, path))
	default:
		return fmt.Errorf("unsupported format for full: %s", format)
	}
}

func wrapExportError(label string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("failed to export %s: %w", label, err)
}

func resolveExportPath(cfg appconfig.RuntimeConfig, format string) (string, error) {
	defaultPath := appconfig.BuildOutputPath(cfg, "."+format)
	if !cfg.UseDialog {
		return defaultPath, nil
	}

	title := fmt.Sprintf("Save %s file", strings.ToUpper(format))
	path, err := appconfig.ChooseSavePath(defaultPath, title)
	if err != nil {
		return "", fmt.Errorf("failed to open save dialog for %s: %w", format, err)
	}
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("save dialog cancelled for %s", format)
	}
	return path, nil
}

func wrapFetchError(prefix string, err error) error {
	if err == nil {
		return nil
	}

	message := fmt.Sprintf("%s: %v", prefix, err)
	if strings.Contains(strings.ToLower(err.Error()), "failed to fetch scraping") {
		message += "\ncheck USER_ID and PASSWORD in .usersetting"
	}
	return fmt.Errorf("%s", message)
}

func defaultFormats(mode string) string {
	if mode == "full" {
		return "json,csv,xlsx,ics"
	}
	return "json,csv,xlsx"
}

func parseFormats(mode string, raw string) ([]string, error) {
	allowed := []string{"json", "csv", "xlsx"}
	if mode == "full" {
		allowed = append(allowed, "ics")
	}

	seen := map[string]bool{}
	var formats []string
	for _, part := range strings.Split(raw, ",") {
		format := strings.TrimSpace(strings.ToLower(part))
		if format == "" {
			continue
		}
		if !slices.Contains(allowed, format) {
			return nil, fmt.Errorf("unsupported format for %s: %s", mode, format)
		}
		if seen[format] {
			continue
		}
		seen[format] = true
		formats = append(formats, format)
	}

	if len(formats) == 0 {
		return nil, fmt.Errorf("no export format specified")
	}
	return formats, nil
}
