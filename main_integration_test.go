// +build integration

package main

import (
	"flag"
	"fmt"
	// "github.com/stretchr/testify/assert"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	flag.Parse()
	shell("pwd")
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

func TestIntegrationGetCustomerId(t *testing.T) {
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
			fmt.Println("reveived timeout")
			return
		}
	}

}
