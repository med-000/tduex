package appconfig

import (
	"os"
	"path/filepath"
)

func DefaultSettingPath() string {
	return filepath.Join(configDir(), ".setting")
}

func DefaultUserSettingPath() string {
	return filepath.Join(configDir(), ".usersetting")
}

func LegacySettingPath() string {
	return ".setting"
}

func LegacyUserSettingPath() string {
	return ".usersetting"
}

func configDir() string {
	dir, err := os.UserConfigDir()
	if err == nil && dir != "" {
		return filepath.Join(dir, "tduex")
	}

	home, err := os.UserHomeDir()
	if err == nil && home != "" {
		return filepath.Join(home, ".config", "tduex")
	}

	return "."
}
