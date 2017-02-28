package util

import (
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
)

// GetConfDirs return the ordered configuration directories.
func GetConfDirs() ([]string, error) {
	var dirs []string
	home, err := homedir.Dir()
	if err != nil {
		return dirs, err
	}
	user := filepath.Join(home, ".config/mirach")
	sys := "/etc/mirach/"
	dirs = append(dirs, ".", user, sys)
	return dirs, nil
}
