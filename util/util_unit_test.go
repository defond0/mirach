// +build unit

package util

import (
	"os"
	"path/filepath"
	"testing"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	// Set util to use the test filesystem.
	Fs = testFs
	writeTestCerts()
	os.Exit(m.Run())
}

func TestGetCA(t *testing.T) {
	assert := assert.New(t)
	correct, err := afero.ReadFile(osFs, "../test_resources/ca.pem")
	ca, err := GetCA([]string{"/etc/mirach/"})
	assert.Nil(err)
	assert.Equal(correct, ca)
}

func TestFindInDirs(t *testing.T) {
	assert := assert.New(t)
	home, err := homedir.Dir()
	assert.Nil(err)
	dirs := []string{"/etc/mirach/", filepath.Join(home, ".config/mirach/")}
	path, err := FindInDirs("ca.pem", dirs)
	assert.Nil(err)
	assert.Equal("/etc/mirach/ca.pem", path, "found in etc")
	path, err = FindInDirs(filepath.Join("customer", "keys", "private.pem.key"), dirs)
	assert.Nil(err)
	assert.Equal(filepath.Join(home, ".config/mirach/customer/keys/private.pem.key"), path, "found in home")
}
