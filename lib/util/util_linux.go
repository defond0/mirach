package util

import (
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
)

// GetConfDirs return the ordered configuration directories.
func GetConfDirs() map[string]string {
	home, _ := homedir.Dir()
	return map[string]string{
		"cur":  ".",
		"user": filepath.Join(home, ".config/mirach"),
		"sys":  "/etc/mirach/",
	}
}
