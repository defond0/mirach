package mirachlib

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"reflect"
	"time"

	"gitlab.eng.cleardata.com/dash/mirach/util"

	"github.com/google/uuid"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/theherk/viper"
)

// MaxMQTTDataSize defines the maximum size of chunks to be sent over MQTT.
const MaxMQTTDataSize = 88000

// MaxChunkedSize defines the threshold after which presigned urls will be used.
const MaxChunkedSize = 512000

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
	Hash string `json:"hash"`
}

type chunksMsg struct {
	mqttMsg
	NumChunks int    `json:"num_chunks"`
	ChunksID  string `json:"chunks_id"`
}

type dataMsg struct {
	mqttMsg
	Data json.RawMessage `json:"data"`
}

type putHTTPMsg struct {
	mqttMsg
	URL string `json:"url"`
}

type getURLMsg struct {
	URL string `json:"url"`
}

// Run will run an external plugin and publishes its results.
func (p *ExternalPlugin) Run(asset *Asset) func() {
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
		if err := SendData([]byte(d), label, asset); err != nil {
			jww.ERROR.Println(err)
		}
	}
}

// Run will run an internal function and publish its results.
func (p *InternalPlugin) Run(asset *Asset) func() {
	f := p.StrFunc
	label := p.Label
	t := p.Type
	return func() {
		defer func() {
			if r := recover(); r != nil {
				if reflect.TypeOf(r).String() == "plugin.Exception" {
					jww.TRACE.Println(r)
					return
				}
				jww.ERROR.Println(r)
			}
		}()
		jww.INFO.Printf("Running internal plugin: %s", label)
		d := f()
		if err := SendData([]byte(d), t, asset); err != nil {
			jww.ERROR.Println(err)
		}
	}
}

// SendData sends data using one of a few methods to an MQTT broker.
func SendData(b []byte, t string, asset *Asset) error {
	custID := viper.GetString("customer.id")
	assetID := viper.GetString("asset.id")
	var err error
	h := md5.Sum(b)
	hash := hex.EncodeToString(h[:])
	msg := mqttMsg{Type: t, Hash: hash}
	var msgB []byte
	switch {
	case len(b) >= MaxChunkedSize:
		url, err := PutData(b, asset)
		m := putHTTPMsg{msg, url}
		msgB, err = json.Marshal(m)
		if err != nil {
			return err
		}
	case len(b) >= MaxMQTTDataSize:
		n, id, err := SendChunks(b, asset)
		if err != nil {
			return err
		}
		m := chunksMsg{msg, n, id}
		msgB, err = json.Marshal(m)
		if err != nil {
			return err
		}
	default:
		m := dataMsg{msg, json.RawMessage(string(b))}
		msgB, err = json.Marshal(m)
		if err != nil {
			return err
		}
	}
	path := fmt.Sprintf("mirach/data/%s/%s", custID, assetID)
	if err := PubWait(asset.client, path, msgB); err != nil {
		return err
	}
	return nil
}

// SendChunks splits a byte slice and return the number of chunks, ID of
// the chunk group, hex digest of the full byte slice's md5 hash, or any
// errors generated along the way.
func SendChunks(b []byte, asset *Asset) (int, string, error) {
	id := fmt.Sprintf("%s", uuid.New())
	var n int
	splits, err := util.SplitAt(b, MaxChunkedSize)
	if err != nil {
		return 0, "", err
	}
	for i, split := range splits {
		path := fmt.Sprintf("mirach/chunk/%s-%d", id, i)
		if err := PubWait(asset.client, path, split); err != nil {
			return 0, "", err
		}
		n++
	}
	return n, id, nil
}

// PutData Gets presigned url and put data to it. Return string of url it has
// been put to
func PutData(b []byte, asset *Asset) (string, error) {
	client := &http.Client{}
	presigned, err := GetPutUrl(asset)
	if err != nil {
		jww.ERROR.Println(err)
		return "", err
	}
	req, err := http.NewRequest("PUT", presigned.URL, bytes.NewBuffer(b))
	if err != nil {
		jww.ERROR.Println(err)
		return "", err
	}
	if _, err = client.Do(req); err != nil {
		jww.ERROR.Println(err)
		return "", err
	}
	return presigned.URL, nil
}

// GetPutUrl will return a presigned url msg or error
func GetPutUrl(asset *Asset) (getURLMsg, error) {
	if err := asset.CycleUrlChannel(); err != nil {
		return getURLMsg{}, err
	}
	custID := viper.GetString("customer.id")
	assetID := viper.GetString("asset.id")
	path := fmt.Sprintf("mirach/url/put/%s/%s", custID, assetID)
	pubToken := asset.client.Publish(path, 1, false, "")
	pubToken.Wait()
	timeoutCh := util.Timeout(10 * time.Second)
	select {
	case res := <-asset.urlChan:
		return res, nil
	case <-timeoutCh:
		return getURLMsg{}, errors.New("Timeout receiving presigned url")
	}
}
