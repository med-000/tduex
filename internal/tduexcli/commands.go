package tduexcli

import (
	"fmt"
	"strings"

	"github.com/med-000/tduscheexport/internal/appconfig"
)

func runClasses(args []string) error {
	cfg, err := parseRuntimeConfig("classes", args)
	if err != nil {
		return err
	}

	svc, err := newService()
	if err != nil {
		return err
	}

	req, _, err := appconfig.BuildRequest(cfg)
	if err != nil {
		return err
	}

	result, err := svc.FetchClassesForExport(req)
	if err != nil {
		return wrapFetchError("failed to fetch class list", err)
	}

	paths := make([]string, 0, len(cfg.Formats))
	for _, format := range cfg.Formats {
		path, err := resolveExportPath(cfg, format)
		if err != nil {
			return err
		}
		if err := exportClassesByFormat(svc, result, format, path); err != nil {
			return err
		}
		paths = append(paths, path)
	}

	fmt.Printf("exported %d classes to %s\n", len(result.Classes), strings.Join(paths, ", "))
	return nil
}

func runFull(args []string) error {
	cfg, err := parseRuntimeConfig("full", args)
	if err != nil {
		return err
	}

	svc, err := newService()
	if err != nil {
		return err
	}

	req, _, err := appconfig.BuildRequest(cfg)
	if err != nil {
		return err
	}

	result, err := svc.FetchFullForExport(req)
	if err != nil {
		return wrapFetchError("failed to fetch full data", err)
	}

	paths := make([]string, 0, len(cfg.Formats))
	for _, format := range cfg.Formats {
		path, err := resolveExportPath(cfg, format)
		if err != nil {
			return err
		}
		if err := exportFullByFormat(svc, result, format, path); err != nil {
			return err
		}
		paths = append(paths, path)
	}

	fmt.Printf("exported %d classes to %s\n", len(result.Classes), strings.Join(paths, ", "))
	return nil
}
