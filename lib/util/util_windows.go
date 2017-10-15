package util

import (
	"os"
	"path/filepath"
)

// GetConfDirs return the ordered configuration directories.
func GetConfDirs() map[string]string {
	return map[string]string{
		"cur":  ".",
		"user": filepath.Join(os.Getenv("LocalAppData"), "mirach"),
		"sys":  filepath.Join(os.Getenv("ProgramData"), "mirach"),
	}
}
