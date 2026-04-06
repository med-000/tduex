package main

import (
	"fmt"
	"os"

	"github.com/med-000/tduscheexport/pkg/appconfig"
	"github.com/med-000/tduscheexport/pkg/logger"
	"github.com/med-000/tduscheexport/pkg/service"
)

const (
	targetYear   = 2025
	targetTerm   = 1
	targetDay    = 0
	targetPeriod = 0
)

func main() {
	if err := appconfig.LoadDotEnv(".env"); err != nil {
		fmt.Fprintf(os.Stderr, "failed to load .env: %v\n", err)
		os.Exit(1)
	}

	req, outputPath, err := appconfig.BuildRequest(appconfig.RuntimeConfig{
		Year:   targetYear,
		Term:   targetTerm,
		Day:    targetDay,
		Period: targetPeriod,
		Mode:   "classes",
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	serviceLogger, err := logger.NewServiceLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	svc := service.NewService(serviceLogger)
	result, err := svc.FetchClassesAndExportJSON(req, outputPath)
	if err != nil {
		serviceLogger.Error.Printf("failed to export class list json: %v", err)
		os.Exit(1)
	}

	fmt.Printf("exported %d classes to %s\n", len(result.Classes), outputPath)
}
