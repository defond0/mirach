package mirachlib

import (
	"encoding/json"
	"fmt"
	"os/exec"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

// ExternalPlugin is a regularly run command that collects data.
type ExternalPlugin struct {
	Cmd      string `json:"cmd"`
	Label    string `json:"label"`
	Schedule string `json:"schedule"`
}

// InternalPlugin is a regularly run function that collects data.
type InternalPlugin struct {
	Label    string
	Schedule string
	StrFunc  func() string
	Type     string
}

type result struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

// Run will run an external plugin and publishes its results.
func (p *ExternalPlugin) Run(c MQTT.Client) func() {
	return func() {
		jww.INFO.Printf("Running external plugin: %s", p.Label)
		cmd := exec.Command(p.Cmd)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			jww.ERROR.Println(err)
		}
		if err := cmd.Start(); err != nil {
			jww.ERROR.Println(err)
		}
		res := result{Type: p.Label}
		if err := json.NewDecoder(stdout).Decode(&res.Data); err != nil {
			jww.ERROR.Println(err)
		}
		err = cmd.Wait()
		custID := viper.GetString("customer.id")
		assetID := viper.GetString("asset.id")
		mes, err := json.Marshal(res)
		if err != nil {
			jww.ERROR.Println(err)
		}
		path := fmt.Sprintf("mirach/data/%s/%s", custID, assetID)
		token := c.Publish(path, 1, false, string(mes))
		token.Wait()
	}
}

// Run will run an internal function and publish its results.
func (p *InternalPlugin) Run(c MQTT.Client) func() {
	return func() {
		jww.INFO.Printf("Running internal plugin: %s", p.Label)
		res := result{Type: p.Type}
		if err := json.Unmarshal([]byte(p.StrFunc()), &res.Data); err != nil {
			jww.ERROR.Println(err)
		}
		custID := viper.GetString("customer.id")
		assetID := viper.GetString("asset.id")
		mes, err := json.Marshal(res)
		if err != nil {
			jww.ERROR.Println(err)
		}
		path := fmt.Sprintf("mirach/data/%s/%s", custID, assetID)
		token := c.Publish(path, 1, false, string(mes))
		token.Wait()
	}
}
