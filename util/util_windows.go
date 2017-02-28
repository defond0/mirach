package util

import (
	"os"
	"path/filepath"
)

// GetConfDirs return the ordered configuration directories.
func GetConfDirs() ([]string, error) {
	var dirs []string
	user := filepath.Join(os.Getenv("LocalAppData"), "mirach")
	sys := filepath.Join(os.Getenv("CommonAppData"), "mirach")
	dirs = append(dirs, ".", user, sys)
	return dirs, nil
}
