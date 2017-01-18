// +build integration

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
	"time"
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
	// shell("rm", "config.yaml")

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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	cmd := exec.CommandContext(ctx, "./mirach")
	stdout, err := cmd.StdoutPipe()
	err := cmd.Start()
	assert.Nil(err)
	select {
	case <-time.After(40 * time.Second):
		fmt.Println("overslept")
	case <-ctx.Done():
		fmt.Println(ctx.Err()) // prints "context deadline exceeded"
	}
	cancel()
	// Even though ctx should have expired already, it is good
	// practice to call its cancelation function in any case.
	// Failure to do so may keep the context and its parent alive
	// longer than necessary.
	// assert := assert.New(t)
	// timeChan := Timeout(10 * time.Second)
	// running := false
	// for {
	// 	if !running {
	// 		fmt.Println("main begining")
	// 		running = true
	// 		go main()
	// 	}
	// 	var timeout = <-timeChan
	// 	if timeout {
	// 		fmt.Println("mirach has run for allotted time interupting and, checking assertions")
	// 		err := viper.ReadInConfig()
	// 		// assert we read config correctly
	// 		assert.Nil(err)
	// 		// assert we received correct customer_id and wrote it to config
	// 		assert.Equal(viper.GetString("customer.id"), "00000666")
	// 		priv, err := ioutil.ReadFile("/etc/mirach/asset/keys/private.pem.key")
	// 		// assert we read file w/o error
	// 		assert.Nil(err)
	// 		// assert we wrote a private key
	// 		assert.NotEmpty(priv)
	// 		ca, err := ioutil.ReadFile("/etc/mirach/asset/keys/ca.pem.crt")
	// 		// assert we read file w/o error
	// 		assert.Nil(err)
	// 		// assert we wrote a cert
	// 		assert.NotEmpty(ca)
	// 		break
	// 	}
	// }
}

//Attempt to register with customer number 00006913(not it's own)
func TestIntegrationMainEvilListener(t *testing.T) {
	cleanup_asset_certs()
	load_evil_test_config()
	assert := assert.New(t)
	go func() {
		time.Sleep(60 * time.Second)
		assert.Fail("mirach did not exit with evil config in 60 seconds")
	}()
	defer func() {
		fmt.Println("mirach has exited, checking assertions to ensure that evil asset not created")
		// assert we used our evil config
		assert.Equal("00006913", viper.GetString("customer.id"))
		priv, err := ioutil.ReadFile("/etc/mirach/asset/keys/private.pem.key")
		assert.NotNil(err)
		assert.Empty(priv)
		ca, err := ioutil.ReadFile("/etc/mirach/asset/keys/ca.pem.crt")
		assert.NotNil(err)
		assert.Empty(ca)

	}()
	_ = shell("go", "run", "main.go")
}
