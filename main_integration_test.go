// +build integration

package main

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
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	flag.Parse()
	os.Args = []string{"-v"}

	load_test_config()
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
	})
	//grab config from test resources
	//grab test customer certs from test resources
	shell("cp", "-r", "test_resources/customer", ".")
	shell("cp", "test_resources/ca.pem", ".")

	//run tests
	run := m.Run()

	//cleanup config
	shell("rm", "config.yaml")

	//cleanup customer-certs
	shell("rm", "-r", "./customer")
	shell("rm", "./ca.pem")
	cleanup_asset_certs()
	os.Exit(run)
}

func load_test_config() {
	shell("cp", "test_resources/config.yaml", ".")
}

func load_evil_test_config() {
	shell("cp", "test_resources/eve_config.yaml", "config.yaml")
}

func cleanup_asset_certs() {
	shell("rm", "-r", "/etc/mirach/asset")
}

func shell(cmd string, args ...string) []byte {
	out, err := exec.Command(cmd, args...).Output()
	if err != nil {
		fmt.Println(err)
	}
	if string(out) != "" && string(out) != " " {
		fmt.Println(string(out))
	}
	return out
}

//Attempt to register with customer number 00006913(not it's own)
func TestIntegrationMainRegistration(t *testing.T) {
	assert := assert.New(t)
	cleanup_asset_certs()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "./mirach")
	stdoutpipe, _ := cmd.StdoutPipe()
	stdoutscanner := bufio.NewScanner(stdoutpipe)
	if err := cmd.Start(); err != nil {
		fmt.Println(err)
	}
	for stdoutscanner.Scan() {
		scan := stdoutscanner.Text()
		if scan == "mirach entered running state; plugins loaded" {
			err := viper.ReadInConfig()
			// assert we read config correctly
			assert.Nil(err)
			// assert we received correct customer_id and wrote it to config
			assert.Equal(viper.GetString("customer.id"), "00000666")
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
	cleanup_asset_certs()
	load_evil_test_config()
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	assert := assert.New(t)
	cmd := exec.CommandContext(ctx, "./mirach")
	stdoutpipe, _ := cmd.StdoutPipe()
	stdoutscanner := bufio.NewScanner(stdoutpipe)
	if err := cmd.Start(); err != nil {
		fmt.Println(err)
	}
	fmt.Println("bout to scan")
	for stdoutscanner.Scan() {
		scan := stdoutscanner.Text()
		fmt.Println(scan)
		if scan == "mirach entered running state; plugins loaded" {
			cancel()
			// assert we used our evil config
			assert.Equal("00006913", viper.GetString("customer.id"))
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
