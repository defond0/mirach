package util

import (
	"path/filepath"

	"github.com/theherk/winpath"
)

// GetConfDirs return the ordered configuration directories.
func GetConfDirs() ([]string, error) {
	var dirs []string
	appData, _ := winpath.LocalAppData()
	comAppData, _ := winpath.CommonAppData()
	user := filepath.Join(appData, "mirach")
	sys := filepath.Join(comAppData, "mirach")
	dirs = append(dirs, ".", user, sys)
	return dirs, nil
}
