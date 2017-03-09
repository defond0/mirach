package mirachlib

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"gitlab.eng.cleardata.com/dash/mirach/util"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/theherk/viper"
)

// ChunkSize defines the maximum size of chunks to be sent over MQTT.
const ChunkSize = 120000

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

type mqttMsg struct {
	Type string `json:"type"`
}

type chunksMsg struct {
	mqttMsg
	Chunks []string `json:"chunks"`
}

type dataMsg struct {
	mqttMsg
	Data json.RawMessage `json:"data"`
}

type urlMsg struct {
	mqttMsg
	URL string `json:"url"`
}

// Run will run an external plugin and publishes its results.
func (p *ExternalPlugin) Run(c mqtt.Client) func() {
	cmd := p.Cmd
	label := p.Label
	return func() {
		jww.INFO.Printf("Running external plugin: %s", label)
		cmd := exec.Command(cmd)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			jww.ERROR.Println(err)
		}
		if err := cmd.Start(); err != nil {
			jww.ERROR.Println(err)
		}
		var d []byte
		if err := json.NewDecoder(stdout).Decode(&d); err != nil {
			jww.ERROR.Println(err)
		}
		err = cmd.Wait()
		if err := SendData([]byte(d), c, label); err != nil {
			jww.ERROR.Println(err)
		}
	}
}

// Run will run an internal function and publish its results.
func (p *InternalPlugin) Run(c mqtt.Client) func() {
	f := p.StrFunc
	label := p.Label
	t := p.Type
	return func() {
		jww.INFO.Printf("Running internal plugin: %s", label)
		d := f()
		if err := SendData([]byte(d), c, t); err != nil {
			jww.ERROR.Println(err)
		}
	}
}

		if err != nil {
			jww.ERROR.Println(err)
		}
		path := fmt.Sprintf("mirach/data/%s/%s", custID, assetID)
		token := c.Publish(path, 1, false, string(mes))
		token.Wait()
	}
}
