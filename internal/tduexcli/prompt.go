package tduexcli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/med-000/tduscheexport/internal/appconfig"
)

func runInteractive() error {
	mode, err := promptMode()
	if err != nil {
		return err
	}

	year, err := promptValidatedInt("year", 2025, appconfig.ValidateYear)
	if err != nil {
		return err
	}

	term, err := promptValidatedInt("term", 1, appconfig.ValidateTerm)
	if err != nil {
		return err
	}

	day, err := promptValidatedInt("day (0=all, 1=Mon ... 7=Sun)", 0, appconfig.ValidateDay)
	if err != nil {
		return err
	}

	period, err := promptValidatedInt("period (0=all)", 0, appconfig.ValidatePeriod)
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

func promptValidatedInt(label string, defaultValue int, validate func(int) error) (int, error) {
	for {
		value, err := promptInt(label, defaultValue)
		if err != nil {
			return 0, err
		}
		if validate == nil {
			return value, nil
		}
		if err := validate(value); err != nil {
			fmt.Fprintln(os.Stderr, err)
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
