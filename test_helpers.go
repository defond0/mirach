// +build integration unit

package main

import (
	"fmt"
	"os/exec"
)

func cleanup_config() {
	//cleanup config
	shell("rm", "./config.yaml")
}

func cleanup_certs() {
	//cleanup customer-certs
	shell("rm", "-r", "./customer")
	shell("rm", "./ca.pem")

}

func load_test_config() {
	shell("cp", "test_resources/config.yaml", "./config.yaml")
}

func load_evil_test_config() {
	shell("cp", "test_resources/eve_config.yaml", "./config.yaml")
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
