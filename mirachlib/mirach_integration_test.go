// +build integration

package mirachlib

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var vipert = viper.New()

func TestMain(m *testing.M) {
	flag.Parse()
	setup()
	//run tests
	run := m.Run()
	cleanup()
	os.Exit(run)

}

func setup() {
	cleanup_asset_certs()
	load_test_config()
	vipert.SetConfigName("config")
	vipert.SetConfigType("yaml")
	vipert.AddConfigPath("./")
	vipert.WatchConfig()
	vipert.OnConfigChange(func(e fsnotify.Event) {
	})
	//grab config from test resources
	//grab test customer certs from test resources
	shell("cp", "-r", "test_resources/customer", "./customer")
	shell("cp", "test_resources/ca.pem", "./ca.pem")
}

func cleanup() {
	cleanup_config()
	cleanup_certs()
}

func TestInitializeConfigAndLogging(t *testing.T) {
	//TODO: use afero move to unit test
	load_test_config()
	assert := assert.New(t)
	Mirach = &mirachSession{}
	opts.Verbose = []bool{}
	Mirach.initializeConfigAndLogging()
	assert.NotNil(Mirach.getSysConfDir())
	assert.NotNil(Mirach.getUserConfDir())
	assert.Equal(Mirach.getVerbosity(), 0, "Expected vebosity of 0")
	assert.Equal(jww.StdoutThreshold(), jww.LevelError)
	assert.Equal(Mirach.getConfigDirs(), []string{".", Mirach.getUserConfDir(), Mirach.getSysConfDir()})
}

func TestGetCA(t *testing.T) {
	//TODO: use afero move to unit test
	assert := assert.New(t)
	correct, err := ioutil.ReadFile("./ca.pem")
	ca, err := getCA()
	assert.Nil(err)
	assert.Equal(correct, ca)
}

func TestIntegrationMainRegistration(t *testing.T) {
	assert := assert.New(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	defer cleanup_asset_certs()
	cmd := exec.CommandContext(ctx, "./mirach")
	stdoutpipe, _ := cmd.StdoutPipe()
	stdoutscanner := bufio.NewScanner(stdoutpipe)
	if err := cmd.Start(); err != nil {
		fmt.Println(err)
	}
	for stdoutscanner.Scan() {
		scan := stdoutscanner.Text()
		if scan == "mirach entered running state; plugins loaded" {
			// assert we received correct customer_id and wrote it to config
			assert.Equal(vipert.GetString("customer.id"), "00000666")
			priv, err := ioutil.ReadFile("/etc/mirach/asset/keys/private.pem.key")
			// assert we read file w/o error
			assert.Nil(err)
			// assert we wrote a private key
			assert.NotEmpty(priv)
			ca, err := ioutil.ReadFile("/etc/mirach/asset/keys/ca.pem.crt")
			// assert we read file w/o error
			assert.Nil(err)
			// assert we wrote a cert
			assert.NotEmpty(ca)
			cancel()
		}
	}
	select {
	case <-ctx.Done():
		assert.Equal("context canceled", fmt.Sprint(ctx.Err()))
	}
}

//Attempt to register with customer number 00006913(not it's own)
func TestIntegrationMainEvilListener(t *testing.T) {
	load_evil_test_config()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	defer cleanup_asset_certs()
	assert := assert.New(t)
	cmd := exec.CommandContext(ctx, "./mirach")
	stdoutpipe, _ := cmd.StdoutPipe()
	stdoutscanner := bufio.NewScanner(stdoutpipe)
	if err := cmd.Start(); err != nil {
		fmt.Println(err)
	}
	for stdoutscanner.Scan() {
		scan := stdoutscanner.Text()
		if scan == "asset initialization failed" {
			cancel()
			// assert we used our evil config
			assert.Equal("00006913", vipert.GetString("customer.id"))
			priv, err := ioutil.ReadFile("/etc/mirach/asset/keys/private.pem.key")
			assert.NotNil(err)
			assert.Empty(priv)
			ca, err := ioutil.ReadFile("/etc/mirach/asset/keys/ca.pem.crt")
			assert.NotNil(err)
			assert.Empty(ca)
		}
	}
	select {
	case <-ctx.Done():
		assert.Equal("context canceled", fmt.Sprint(ctx.Err()))
	}

}
