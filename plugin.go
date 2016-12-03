package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/golang/glog"
)

type Plugin struct {
	Label    string `json:"label"`
	Cmd      string `json:"cmd"`
	Schedule string `json:"schedule"`
}

type result struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

func RunPlugin(p Plugin, c MQTT.Client) func() {
	return func() {
		glog.Infof("Running plugin: %s", p.Cmd)
		cmd := exec.Command(p.Cmd)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Fatal(err)
		}
		if err := cmd.Start(); err != nil {
			log.Fatal(err)
		}
		var res result
		if err := json.NewDecoder(stdout).Decode(&res); err != nil {
			log.Fatal(err)
		}
		err = cmd.Wait()
		Publish(res, c)
		fmt.Printf("type: %s, data: %s\n", res.Type, res.Data)
	}
}
