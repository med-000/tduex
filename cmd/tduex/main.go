package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/med-000/tduscheexport/pkg/appconfig"
	"github.com/med-000/tduscheexport/pkg/logger"
	"github.com/med-000/tduscheexport/pkg/service"
)

func main() {
	if len(os.Args) < 2 {
		if err := runInteractive(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	switch os.Args[1] {
	case "classes":
		if err := runClasses(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "full":
		if err := runFull(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "-h", "--help", "help":
		printUsage(os.Stdout)
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n\n", os.Args[1])
		printUsage(os.Stderr)
		os.Exit(1)
	}
}

func runInteractive() error {
	mode, err := promptMode()
	if err != nil {
		return err
	}

	year, err := promptInt("year", 2025)
	if err != nil {
		return err
	}

	term, err := promptInt("term", 1)
	if err != nil {
		return err
	}

	day, err := promptInt("day (0=all, 1=Mon ... 7=Sun)", 0)
	if err != nil {
		return err
	}

	period, err := promptInt("period (0=all)", 0)
	if err != nil {
		return err
	}

	formats, err := promptFormats(mode)
	if err != nil {
		return err
	}

	if mode == "classes" {
		return runClasses(buildPromptArgs(year, term, day, period, formats))
	}
	return runFull(buildPromptArgs(year, term, day, period, formats))
}

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
		return fmt.Errorf("failed to fetch class list: %w", err)
	}

	paths := make([]string, 0, len(cfg.Formats))
	for _, format := range cfg.Formats {
		path, err := resolveExportPath(cfg, format)
		if err != nil {
			return err
		}
		switch format {
		case "json":
			if err := svc.ExportCourseJSON(result, path); err != nil {
				return fmt.Errorf("failed to export class list json: %w", err)
			}
		case "csv":
			if err := svc.ExportCourseCSV(result, path); err != nil {
				return fmt.Errorf("failed to export class list csv: %w", err)
			}
		case "xlsx":
			if err := svc.ExportCourseXLSX(result, path); err != nil {
				return fmt.Errorf("failed to export class list xlsx: %w", err)
			}
		default:
			return fmt.Errorf("unsupported format for classes: %s", format)
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
		return fmt.Errorf("failed to fetch full data: %w", err)
	}

	paths := make([]string, 0, len(cfg.Formats))
	for _, format := range cfg.Formats {
		path, err := resolveExportPath(cfg, format)
		if err != nil {
			return err
		}
		switch format {
		case "json":
			if err := svc.ExportFullCourseJSON(result, path); err != nil {
				return fmt.Errorf("failed to export full json: %w", err)
			}
		case "csv":
			if err := svc.ExportFullCourseCSV(result, path); err != nil {
				return fmt.Errorf("failed to export full csv: %w", err)
			}
		case "xlsx":
			if err := svc.ExportFullCourseXLSX(result, path); err != nil {
				return fmt.Errorf("failed to export full xlsx: %w", err)
			}
		case "ics":
			if err := svc.ExportFullCourseICS(result, path); err != nil {
				return fmt.Errorf("failed to export full ics: %w", err)
			}
		default:
			return fmt.Errorf("unsupported format for full: %s", format)
		}
		paths = append(paths, path)
	}

	fmt.Printf("exported %d classes to %s\n", len(result.Classes), strings.Join(paths, ", "))
	return nil
}

func parseRuntimeConfig(mode string, args []string) (appconfig.RuntimeConfig, error) {
	fs := flag.NewFlagSet(mode, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	year := fs.Int("year", 2025, "target year")
	term := fs.Int("term", 1, "target term")
	day := fs.Int("day", 0, "target day (0=all, 1=Mon ... 7=Sun)")
	period := fs.Int("period", 0, "target period (0=all)")
	formatsFlag := fs.String("format", defaultFormats(mode), "export format(s), comma-separated")
	dotenvPath := fs.String("env", ".env", "path to .env file")
	settingPath := fs.String("setting", ".setting", "path to credential setting file")
	useDialog := fs.Bool("dialog", true, "show native save dialog")

	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "usage: tduex %s [options]\n", mode)
	}

	if err := fs.Parse(args); err != nil {
		return appconfig.RuntimeConfig{}, err
	}
	if fs.NArg() != 0 {
		return appconfig.RuntimeConfig{}, fmt.Errorf("unexpected arguments: %v", fs.Args())
	}

	if err := appconfig.LoadOptionalEnv(*settingPath); err != nil {
		return appconfig.RuntimeConfig{}, fmt.Errorf("failed to load %s: %w", *settingPath, err)
	}
	if err := appconfig.LoadOptionalEnv(*dotenvPath); err != nil {
		return appconfig.RuntimeConfig{}, fmt.Errorf("failed to load %s: %w", *dotenvPath, err)
	}
	if err := ensureCredentials(*settingPath); err != nil {
		return appconfig.RuntimeConfig{}, err
	}

	formats, err := parseFormats(mode, *formatsFlag)
	if err != nil {
		return appconfig.RuntimeConfig{}, err
	}

	return appconfig.RuntimeConfig{
		Year:      *year,
		Term:      *term,
		Day:       *day,
		Period:    *period,
		Mode:      mode,
		Formats:   formats,
		UseDialog: *useDialog,
	}, nil
}

func newService() (*service.Service, error) {
	serviceLogger, err := logger.NewServiceLogger()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}
	return service.NewService(serviceLogger), nil
}

func printUsage(out *os.File) {
	fmt.Fprintln(out, "usage: tduex [classes|full] [options]")
	fmt.Fprintln(out, "example: tduex full -year 2025 -term 1 -format json,csv,ics")
}

func ensureCredentials(settingPath string) error {
	userID := strings.TrimSpace(os.Getenv("USER_ID"))
	password := strings.TrimSpace(os.Getenv("PASSWORD"))
	if userID != "" && password != "" {
		return nil
	}

	fmt.Println("USER_ID or PASSWORD was not found. They will be saved to .setting.")

	if userID == "" {
		value, err := promptString("USER_ID")
		if err != nil {
			return err
		}
		userID = value
	}
	if password == "" {
		value, err := promptString("PASSWORD")
		if err != nil {
			return err
		}
		password = value
	}

	if err := os.Setenv("USER_ID", userID); err != nil {
		return err
	}
	if err := os.Setenv("PASSWORD", password); err != nil {
		return err
	}

	if err := appconfig.PersistCredentials(settingPath, userID, password); err != nil {
		return fmt.Errorf("failed to write %s: %w", settingPath, err)
	}

	return nil
}

func promptMode() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("export unit [classes/full]: ")
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		mode := strings.TrimSpace(strings.ToLower(line))
		switch mode {
		case "classes", "class":
			return "classes", nil
		case "full", "all":
			return "full", nil
		}

		fmt.Fprintln(os.Stderr, "please input 'classes' or 'full'")
	}
}

func promptInt(label string, defaultValue int) (int, error) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s [%d]: ", label, defaultValue)
		line, err := reader.ReadString('\n')
		if err != nil {
			return 0, err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			return defaultValue, nil
		}

		value, err := strconv.Atoi(line)
		if err != nil {
			fmt.Fprintln(os.Stderr, "please input a number")
			continue
		}
		return value, nil
	}
}

func promptFormats(mode string) ([]string, error) {
	raw, err := promptStringWithDefault("format (comma-separated)", defaultFormats(mode))
	if err != nil {
		return nil, err
	}
	return parseFormats(mode, raw)
}

func promptString(label string) (string, error) {
	return promptStringWithDefault(label, "")
}

func promptStringWithDefault(label string, defaultValue string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	for {
		if defaultValue == "" {
			fmt.Printf("%s: ", label)
		} else {
			fmt.Printf("%s [%s]: ", label, defaultValue)
		}
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		value := strings.TrimSpace(line)
		if value == "" && defaultValue != "" {
			return defaultValue, nil
		}
		if value == "" {
			fmt.Fprintln(os.Stderr, "please input a value")
			continue
		}
		return value, nil
	}
}

func buildPromptArgs(year int, term int, day int, period int, formats []string) []string {
	args := []string{
		"-year", strconv.Itoa(year),
		"-term", strconv.Itoa(term),
		"-day", strconv.Itoa(day),
		"-period", strconv.Itoa(period),
	}
	if len(formats) > 0 {
		args = append(args, "-format", strings.Join(formats, ","))
	}
	return args
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
