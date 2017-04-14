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
	"strings"
	"time"

	"gitlab.eng.cleardata.com/dash/mirach/util"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/theherk/viper"
)

// ChunkSize defines the maximum size of chunks to be sent over MQTT.
const ChunkSize = 88000

// MaxChunkSize defines the threshold after which s3 presigned urls will be used.
const MaxChunkSize = 512000

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
	NumChunks int    `json:"num_chunks"`
	ChunksID  string `json:"chunks_id"`
	ChunksSum string `json:"chunks_sum"`
}

type dataMsg struct {
	mqttMsg
	Data json.RawMessage `json:"data"`
}

type s3Msg struct {
	mqttMsg
	Hash string `json:"hash"`
	Key  string `json:"s3_key"`
}

type urlMsg struct {
	URL string `json:"url"`
	Key string `json:"key"`
}

// Run will run an external plugin and publishes its results.
func (p *ExternalPlugin) Run(c mqtt.Client, urlChan chan urlMsg) func() {
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
		if err := SendData([]byte(d), c, label, urlChan); err != nil {
			jww.ERROR.Println(err)
		}
	}
}

// Run will run an internal function and publish its results.
func (p *InternalPlugin) Run(c mqtt.Client, urlChan chan urlMsg) func() {
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
		if err := SendData([]byte(d), c, t, urlChan); err != nil {
			jww.ERROR.Println(err)
		}
	}
}

// SendData sends data using one of a few methods to an MQTT broker.
func SendData(b []byte, c mqtt.Client, t string, urlChan chan urlMsg) error {
	custID := viper.GetString("customer.id")
	assetID := viper.GetString("asset.id")
	var err error
	msg := mqttMsg{Type: t}
	var msgB []byte
	switch {
	case len(b) >= MaxChunkSize:
		key, hash, err := PutData(b, c, urlChan)
		m := s3Msg{msg, hash, key}
		msgB, err = json.Marshal(m)
		if err != nil {
			return err
		}
	case len(b) >= ChunkSize:
		n, id, sum, err := SendChunks(b, c)
		if err != nil {
			return err
		}
		m := chunksMsg{msg, n, id, sum}
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
	if err := PubWait(c, path, msgB); err != nil {
		return err
	}
	return nil
}

// SendChunks splits a byte slice and return the number of chunks, ID of
// the chunk group, hex digest of the full byte slice's md5 hash, or any
// errors generated along the way.
func SendChunks(b []byte, c mqtt.Client) (int, string, string, error) {
	id := fmt.Sprintf("%s", uuid.New())
	h := md5.Sum(b)
	sum := hex.EncodeToString(h[:])
	var n int
	splits, err := util.SplitAt(b, ChunkSize)
	if err != nil {
		return 0, "", "", err
	}
	for i, split := range splits {
		path := fmt.Sprintf("mirach/chunk/%s-%d", id, i)
		if err := PubWait(c, path, split); err != nil {
			return 0, "", "", err
		}
		n++
	}
	return n, id, sum, nil
}

// PutData Gets presigned url and put data to it. Return string of s3 key it has
// been put to as well as the hash of the bytes
func PutData(b []byte, c mqtt.Client, urlChan chan urlMsg) (string, string, error) {
	client := &http.Client{}
	presigned, err := GetPresignedUrl(c, urlChan)
	if err != nil {
		jww.ERROR.Println(err)
		return "", "", err
	}
	req, err := http.NewRequest("PUT", presigned.URL, bytes.NewBuffer(b))
	if err != nil {
		jww.ERROR.Println(err)
		return "", "", err
	}
	res, err := client.Do(req)
	if err != nil {
		jww.ERROR.Println(err)
		return "", "", err
	}
	h := md5.Sum(b)
	sum := hex.EncodeToString(h[:])
	etag := strings.Trim(res.Header["Etag"][0], "\"")
	if sum == etag {
		return presigned.Key, etag, nil
	} else {
		e := errors.New("etag of uploaded bytes do match calculated hash")
		panic(e)
	}
}

// GetPresignedUrl will return a presigned url msg or error
func GetPresignedUrl(c mqtt.Client, urlChan chan urlMsg) (urlMsg, error) {
	custID := viper.GetString("customer.id")
	assetID := viper.GetString("asset.id")
	path := fmt.Sprintf("mirach/s3/put/%s/%s", custID, assetID)
	pubToken := c.Publish(path, 1, false, "")
	pubToken.Wait()
	timeoutCh := util.Timeout(10 * time.Second)
	select {
	case res := <-urlChan:
		return res, nil
	case <-timeoutCh:
		return urlMsg{}, errors.New("Timeout receiving presigned url")
	}
}
