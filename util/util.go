package util

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/afero"
	jww "github.com/spf13/jwalterweatherman"
)

// Fs is the afero filesystem used in some util functions.
// It can be set to another filesystem by calling SetFs, but defaults to afero.OsFs.
var Fs = afero.NewOsFs()

// Exists is a simple wrapper around afero.Exists.
func Exists(path string) (bool, error) {
	return afero.Exists(Fs, path)
}

// FindInDirs looks for a filename in configured directories,
// and returns the first matching file path.
func FindInDirs(fname string, dirs []string) (string, error) {
	jww.INFO.Printf("searching for %s in %s", fname, dirs)
	for _, d := range dirs {
		fpath := filepath.Join(d, fname)
		if b, _ := afero.Exists(Fs, fpath); b {
			return fpath, nil
		}
	}

	jww.INFO.Printf("unable to find %s in %s", fname, dirs)
	return "", fmt.Errorf("unable to find %s in %s", fname, dirs)
}

// ForceWrite forcibly writes a string to a given filepath.
func ForceWrite(path string, contents string) error {
	if err := Fs.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := Fs.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(contents)
	if err != nil {
		return err
	}
	return nil
}

// GetCA returns the certificate authority pem bytes.
func GetCA(dirs []string) ([]byte, error) {
	caPath, err := FindInDirs("ca.pem", dirs)
	if err != nil {
		return nil, err
	}
	ca, err := afero.ReadFile(Fs, caPath)
	if err != nil {
		return nil, err
	}
	return ca, nil
}

// ReadFile is a simple wrapper around afero.ReadFile.
// afero.ReadFile is an implementation of the ReadFile interface from ioutil,
// but operates on the afero filesystem.
func ReadFile(path string) ([]byte, error) {
	return afero.ReadFile(Fs, path)
}

// SetFs sets the afero filesystem.
// Not setting this will us OsFs by default.
func SetFs(fs afero.Fs) {
	Fs = fs
}

// Timeout starts a go routine which writes true to the given channel
// after the given time.
func Timeout(d time.Duration) <-chan bool {
	ch := make(chan bool, 1)
	go func() {
		time.Sleep(d)
		ch <- true
	}()
	return ch
}
