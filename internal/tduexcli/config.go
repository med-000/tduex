package tduexcli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/med-000/tduscheexport/internal/appconfig"
	"github.com/med-000/tduscheexport/internal/logger"
	"github.com/med-000/tduscheexport/internal/service"
)

func parseRuntimeConfig(mode string, args []string) (appconfig.RuntimeConfig, error) {
	fs := flag.NewFlagSet(mode, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	year := fs.Int("year", 2025, "target year")
	term := fs.Int("term", 1, "target term")
	day := fs.Int("day", 0, "target day (0=all, 1=Mon ... 7=Sun)")
	period := fs.Int("period", 0, "target period (0=all)")
	formatsFlag := fs.String("format", defaultFormats(mode), "export format(s), comma-separated")
	dotenvPath := fs.String("env", ".env", "path to .env file")
	settingPath := fs.String("setting", appconfig.DefaultSettingPath(), "path to app setting file")
	userSettingPath := fs.String("usersetting", appconfig.DefaultUserSettingPath(), "path to user credential setting file")
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
	if *settingPath != appconfig.LegacySettingPath() {
		if err := appconfig.LoadOptionalEnv(appconfig.LegacySettingPath()); err != nil {
			return appconfig.RuntimeConfig{}, fmt.Errorf("failed to load %s: %w", appconfig.LegacySettingPath(), err)
		}
	}
	if err := appconfig.LoadOptionalEnv(*userSettingPath); err != nil {
		return appconfig.RuntimeConfig{}, fmt.Errorf("failed to load %s: %w", *userSettingPath, err)
	}
	if *userSettingPath != appconfig.LegacyUserSettingPath() {
		if err := appconfig.LoadOptionalEnv(appconfig.LegacyUserSettingPath()); err != nil {
			return appconfig.RuntimeConfig{}, fmt.Errorf("failed to load %s: %w", appconfig.LegacyUserSettingPath(), err)
		}
	}
	if err := appconfig.LoadOptionalEnv(*dotenvPath); err != nil {
		return appconfig.RuntimeConfig{}, fmt.Errorf("failed to load %s: %w", *dotenvPath, err)
	}
	if err := ensureCredentials(*userSettingPath); err != nil {
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

func ensureCredentials(userSettingPath string) error {
	userID := strings.TrimSpace(os.Getenv("USER_ID"))
	password := strings.TrimSpace(os.Getenv("PASSWORD"))
	if userID != "" && password != "" {
		return nil
	}

	fmt.Println("USER_ID or PASSWORD was not found. They will be saved to .usersetting.")

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

	if err := appconfig.PersistCredentials(userSettingPath, userID, password); err != nil {
		return fmt.Errorf("failed to write %s: %w", userSettingPath, err)
	}

	return nil
}
