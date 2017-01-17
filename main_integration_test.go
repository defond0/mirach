// +build integration

package main

import (
	"flag"
	"fmt"
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

	//grab config from test resources
	shell("cp", "test_resources/config.yaml", ".")

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
	shell("rm", "-r", "/etc/mirach/asset")

	os.Exit(run)
}

func load_test_config() {
	shell("cp", "test_resources/config.yaml", ".")
}

// func clear_test_

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

func TestIntegrationMainRegistration(t *testing.T) {
	assert := assert.New(t)
	timeChan := Timeout(15 * time.Second)
	running := false
	for {
		if !running {
			fmt.Println("main begining")
			running = true
			go main()
		}
		var timeout = <-timeChan
		if timeout {
			fmt.Println("mirach has run for allotted time interupting and, checking assertions")
			viper.SetConfigName("config")
			viper.SetConfigType("yaml")
			viper.AddConfigPath(".")
			err := viper.ReadInConfig()
			assert.Nil(err)
			assert.Equal(viper.GetString("customer.id"), "00000666")
			priv, err := ioutil.ReadFile("/etc/mirach/asset/keys/private.pem.key")
			assert.Nil(err)
			assert.NotEmpty(priv)
			ca, err := ioutil.ReadFile("/etc/mirach/asset/keys/ca.pem.crt")
			assert.Nil(err)
			assert.NotEmpty(ca)
			break
		}
	}
}
