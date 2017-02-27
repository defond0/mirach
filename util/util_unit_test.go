// +build unit

package util

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/theherk/viper"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	// Set util to use the test filesystem.
	SetFs(TestFs)
	WriteTestCerts()
	WriteTestGoodConfig()
	os.Exit(m.Run())
}

func TestFindInDirs(t *testing.T) {
	assert := assert.New(t)
	dirs := []string{"/etc/mirach/"}
	path, err := FindInDirs("ca.pem", dirs)
	assert.Nil(err)
	assert.Equal("/etc/mirach/ca.pem", path, "found in etc")
	path, err = FindInDirs(filepath.Join("customer", "keys", "private.pem.key"), dirs)
	assert.Nil(err)
	assert.Equal("/etc/mirach/customer/keys/private.pem.key", path, "found in etc")
}

func TestGetCA(t *testing.T) {
	assert := assert.New(t)
	correct, err := afero.ReadFile(OSFs, "../test_resources/ca.pem")
	ca, err := GetCA([]string{"/etc/mirach/"})
	assert.Nil(err)
	assert.Equal(correct, ca)
}

func TestGetConfDirs(t *testing.T) {
	assert := assert.New(t)
	dirs, err := GetConfDirs()
	assert.Nil(err)
	assert.Len(dirs, 3, "current, user, and system")
}

func TestGetConfig(t *testing.T) {
	assert := assert.New(t)
	dirs := []string{"/etc/mirach/"}
	conf, err := GetConfig(dirs)
	assert.Nil(err)
	assert.NotNil(conf)
	assert.Equal("00000666-mirach", viper.GetString("asset.id"), "value in read config")
}
