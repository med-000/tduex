package appconfig

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func ChooseSavePath(defaultPath string, title string) (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return chooseSavePathMac(defaultPath, title)
	case "windows":
		return chooseSavePathWindows(defaultPath, title)
	default:
		return defaultPath, nil
	}
}

func chooseSavePathMac(defaultPath string, title string) (string, error) {
	defaultDir := filepath.Dir(defaultPath)
	defaultName := filepath.Base(defaultPath)
	if _, err := os.Stat(defaultDir); err != nil {
		defaultDir = "."
	}

	script := fmt.Sprintf(
		`set chosenFile to choose file name with prompt "%s" default location POSIX file "%s" default name "%s"
POSIX path of chosenFile`,
		escapeAppleScript(title),
		escapeAppleScript(defaultDir),
		escapeAppleScript(defaultName),
	)

	cmd := exec.Command("osascript", "-e", script)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		if isDialogCanceled(stdout.String(), stderr.String(), err) {
			return "", fmt.Errorf("save dialog canceled")
		}
		return "", fmt.Errorf("%v: %s", err, strings.TrimSpace(stderr.String()))
	}

	return strings.TrimSpace(stdout.String()), nil
}

func chooseSavePathWindows(defaultPath string, title string) (string, error) {
	defaultName := filepath.Base(defaultPath)
	initialDir := filepath.Dir(defaultPath)
	extension := strings.TrimPrefix(filepath.Ext(defaultPath), ".")
	if _, err := os.Stat(initialDir); err != nil {
		initialDir = "."
	}

	script := fmt.Sprintf(
		`Add-Type -AssemblyName System.Windows.Forms
$dialog = New-Object System.Windows.Forms.SaveFileDialog
$dialog.Title = '%s'
$dialog.FileName = '%s'
$dialog.InitialDirectory = '%s'
$dialog.DefaultExt = '%s'
$dialog.AddExtension = $true
if ($dialog.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) { Write-Output $dialog.FileName }`,
		escapePowerShell(title),
		escapePowerShell(defaultName),
		escapePowerShell(initialDir),
		escapePowerShell(extension),
	)

	cmd := exec.Command("powershell", "-NoProfile", "-Command", script)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		if isDialogCanceled(stdout.String(), stderr.String(), err) {
			return "", fmt.Errorf("save dialog canceled")
		}
		return "", fmt.Errorf("%v: %s", err, strings.TrimSpace(stderr.String()))
	}

	return strings.TrimSpace(stdout.String()), nil
}

func escapeAppleScript(value string) string {
	return strings.ReplaceAll(value, `"`, `\"`)
}

func escapePowerShell(value string) string {
	return strings.ReplaceAll(value, `'`, `''`)
}

func isDialogCanceled(stdout string, stderr string, err error) bool {
	combined := strings.ToLower(stdout + "\n" + stderr)
	if strings.Contains(combined, "user canceled") {
		return true
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return len(exitErr.Stderr) == 0 && strings.TrimSpace(stdout) == "" && strings.TrimSpace(stderr) == ""
	}

	return false
}
