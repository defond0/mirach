package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/golang/glog"
	"github.com/spf13/viper"
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

// RunPlugin run a plugin and publishes its results.
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
		custID := viper.GetString("customer_id")
		assetID := viper.GetString("asset_id")
		mes, err := json.Marshal(res)
		if err != nil {
			log.Fatal(err)
		}
		path := fmt.Sprintf("mirach/data/%s/%s", custID, assetID)
		token := c.Publish(path, 0, false, string(mes))
		token.Wait()
	}
}
