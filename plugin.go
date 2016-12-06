package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"time"

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
	Type    string `json:"type"`
	Data    string `json:"data"`
	Time    string `json:"time"`
	AssetID string `json:"asset_id"`
	CustID  string `json:"customer_id"`
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
		custID := viper.GetString("customer_id")
		assetID := viper.GetString("asset_id")
		res.AssetID = assetID
		res.CustID = custID
		res.Time = fmt.Sprint(time.Now().Unix())
		mes, err := json.Marshal(res)
		if err != nil {
			log.Fatal(err)
		}
		path := fmt.Sprintf("mirach/data/%s/%s", custID, assetID)
		Publish(string(mes), path, c)
	}
}
