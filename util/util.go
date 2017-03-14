package util

import (
	"fmt"
	"path/filepath"
	"time"
	"unicode/utf8"

	"github.com/spf13/afero"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/theherk/viper"
)

// Fs is the afero filesystem used in some util functions.
// It can be set to another filesystem by calling SetFs, but defaults to afero.OsFs.
var Fs = afero.NewOsFs()

// CheckExceptions checks to see if given error is in the given list of exceptions.
// It returns the string of the known exception if found or the error if not.
func CheckExceptions(err error, exceptions []string) (string, error) {
	for _, s := range exceptions {
		if err.Error() == s {
			return s, nil
		}
	}
	return "", err
}

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

// GetConfig loads the configuration and return the config file used.
func GetConfig(dirs []string) (string, error) {
	viper.Reset()
	viper.SetConfigName("config")
	for _, d := range dirs {
		viper.AddConfigPath(d)
	}
	viper.SetFs(Fs)
	err := viper.ReadInConfig()
	if err != nil {
		return "", fmt.Errorf("Fatal error config file: %s \n", err)
	}
	viper.SetEnvPrefix("mirach")
	viper.AutomaticEnv()
	return viper.ConfigFileUsed(), nil
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

// SplitAt splits a byte slice into a variable number of byte slices no
// larger than a given number of bytes.
// Each byte slice will be of the given size, except the last, which may
// be smaller.
func SplitAt(b []byte, size int) ([][]byte, error) {
	var chunks [][]byte
	for len(b) > size {
		chunks = append(chunks, b[:size])
		b = b[size:]
	}
	if len(b) > 0 {
		chunks = append(chunks, b)
	}
	return chunks, nil
}

// SplitStringAt splits a string into a variable number of strings no larger than
// a given number of bytes.
func SplitStringAt(s string, size int) ([]string, error) {
	var chunks []string
	for len(s) > size {
		i := size
		for i >= size-utf8.UTFMax && !utf8.RuneStart(s[i]) {
			i--
		}
		chunks = append(chunks, s[:i])
		s = s[i:]
	}
	if len(s) > 0 {
		chunks = append(chunks, s)
	}
	return chunks, nil
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
