package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	jww "github.com/spf13/jwalterweatherman"
)

// customOut either outputs feedback or a log message at error level.
func customOut(fbMsg, err interface{}) {
	switch {
	case verbosity > 0:
		if err != nil {
			jww.ERROR.Println(fmt.Sprint(err))
		} else {
			jww.INFO.Println(fmt.Sprint(fbMsg))
		}
	default:
		jww.FEEDBACK.Println(fmt.Sprint(fbMsg))
	}
}

// Check if File / Directory Exists
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// findInDirs looks for a filename in configured directories,
// and returns the first matching file path.
func findInDirs(fname string, dirs []string) (string, error) {
	jww.INFO.Printf("searching for %s in %s", fname, dirs)
	for _, d := range dirs {
		fpath := filepath.Join(d, fname)
		if b, _ := exists(fpath); b {
			return fpath, nil
		}
	}
	return "", fmt.Errorf("unable to find %s in %s", fname, dirs)
}

// ForceWrite forcibly writes a string to a given filepath.
func ForceWrite(path string, contents string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
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

// getCA returns the certificate authority pem bytes.
func getCA() ([]byte, error) {
	caPath, err := findInDirs("ca.pem", configDirs)
	if err != nil {
		return nil, err
	}
	ca, err := ioutil.ReadFile(caPath)
	if err != nil {
		return nil, err
	}
	return ca, nil
}

// showVersion will print the version information.
func showVersion() {
	fmt.Println("mirach " + version)
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
