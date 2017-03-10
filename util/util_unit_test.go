// +build unit

package util

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/theherk/viper"
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

func ExampleSplitAt() {
	b, _ := SplitAt([]byte("abc ⌘ efg"), 5)
	fmt.Println(b[1])
	// Output: [140 152 32 101 102]
}

func TestSplitAt(t *testing.T) {
	in := []byte("⌘⌘ test")
	size := 5
	b0 := []byte{226, 140, 152, 226, 140}
	b1 := []byte{152, 32, 116, 101, 115}
	b2 := []byte{116}
	exp := [][]byte{b0, b1, b2}
	out, err := SplitAt(in, size)
	if err != nil {
		t.Error("how even")
	}
	for x := range out {
		for y := range out[x] {
			if out[x][y] != exp[x][y] {
				t.Errorf("slice split incorrectly: %v", out)
			}
		}
	}
}

func ExampleSplitStringAt() {
	s, _ := SplitStringAt("abc ⌘ efg", 5)
	fmt.Println(s[1] + ", " + s[2])
	// Output: ⌘ e, fg
}

func TestSplitStringAt(t *testing.T) {
	in := "⌘⌘ test"
	size := 5
	exp := []string{"⌘", "⌘ t", "est"}
	out, err := SplitStringAt(in, size)
	if err != nil {
		t.Error("how even")
	}
	for i := range out {
		if out[i] != exp[i] {
			t.Errorf("string split incorrectly: %v", out)
		}
	}
}
