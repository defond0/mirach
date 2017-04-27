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

	"gitlab.eng.cleardata.com/dash/mirach/cron"
	"gitlab.eng.cleardata.com/dash/mirach/plugin/compinfo"
	"gitlab.eng.cleardata.com/dash/mirach/plugin/ebsinfo"
	"gitlab.eng.cleardata.com/dash/mirach/plugin/envinfo"
	"gitlab.eng.cleardata.com/dash/mirach/plugin/pkginfo"
	"gitlab.eng.cleardata.com/dash/mirach/util"

	"github.com/google/uuid"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/theherk/viper"
)

// MaxMQTTDataSize defines the maximum size of chunks to be sent over MQTT.
const MaxMQTTDataSize = 88000

// MaxChunkedSize defines the threshold after which presigned urls will be used.
const MaxChunkedSize = 512000

type chunksMsg struct {
	mqttMsg
	NumChunks int    `json:"num_chunks"`
	ChunksID  string `json:"chunks_id"`
}

// BuiltinPlugin is a regularly run function that collects data.
type BuiltinPlugin struct {
	Plugin  `mapstructure:",squash"`
	StrFunc func() string
}

// CustomPlugin is a regularly run command that collects data.
type CustomPlugin struct {
	Plugin `mapstructure:",squash"`
	Cmd    string
}

type dataMsg struct {
	mqttMsg
	Data json.RawMessage `json:"data"`
}

type getURLMsg struct {
	URL string `json:"url"`
}

type mqttMsg struct {
	Type string `json:"type"`
	Hash string `json:"hash"`
}

// Plugin is a routine used to collect data.
type Plugin struct {
	Disabled  bool
	Label     string
	LoadDelay string `mapstructure:"load_delay"`
	Schedule  string
	Type      string
}

type putHTTPMsg struct {
	mqttMsg
	URL string `json:"url"`
}

var (
	customPlugins  map[string]CustomPlugin
	builtinPlugins map[string]BuiltinPlugin
)

// Run will run custom plugin and publishes its results.
func (p *CustomPlugin) Run(asset *Asset) func() {
	pCmd := p.Cmd
	pLabel := p.Label
	pType := p.Type
	return func() {
		jww.INFO.Printf("Running plugin: %s", pLabel)
		cmd := exec.Command(pCmd)
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
		if err := SendData(d, pType, asset); err != nil {
			jww.ERROR.Println(err)
		}
	}
}

// Run will run an internal function and publish its results.
func (p *BuiltinPlugin) Run(asset *Asset) func() {
	pFunc := p.StrFunc
	pLabel := p.Label
	pType := p.Type
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
		jww.INFO.Printf("Running built-in plugin: %s", pLabel)
		d := pFunc()
		if err := SendData([]byte(d), pType, asset); err != nil {
			jww.ERROR.Println(err)
		}
	}
}

func (p *Plugin) loadPlugin(cron *cron.MirachCron, f func()) {
	if p.Disabled {
		jww.INFO.Printf("plugin disabled, skipping: %s", p.Label)
		return
	}
	delay, err := time.ParseDuration(p.LoadDelay)
	if err != nil {
		if p.LoadDelay != "" {
			msg := "invalid duration: continuing without delay"
			util.CustomOut(msg, err)
		}
	} else {
		jww.INFO.Printf("adding plugin to cron with start delay: %s, %s", p.Label, delay)
		res := make(chan interface{})
		cron.AddFuncDelayed(p.Schedule, f, delay, res)
		successMsg := fmt.Sprintf("added plugin to cron after delay: %s, %s", p.Label, delay)
		errorMsg := fmt.Sprintf("failed to add plugin to cron after delay: %s, %s", p.Label, delay)
		go logResChan(successMsg, errorMsg, res)
		return
	}
	jww.INFO.Printf("adding plugin to cron: %s", p.Label)
	err = cron.AddFunc(p.Schedule, f)
	if err != nil {
		msg := fmt.Sprintf("failed to load plugin %v", p.Label)
		util.CustomOut(msg, err)
	}
}

func getBuiltinPlugins() map[string]BuiltinPlugin {
	if len(builtinPlugins) == 0 {
		builtinPlugins = map[string]BuiltinPlugin{
			"compinfo-docker": {
				Plugin: Plugin{
					Schedule: "@hourly",
					Type:     "compinfo",
				},
				StrFunc: compinfo.GetDockerString,
			},
			"compinfo-load": {
				Plugin: Plugin{
					Schedule: "@every 5m",
					Type:     "compinfo",
				},
				StrFunc: compinfo.GetLoadString,
			},
			"compinfo-sys": {
				Plugin: Plugin{
					Schedule: "@daily",
					Type:     "compinfo",
				},
				StrFunc: compinfo.GetSysString,
			},
			"pkginfo": {
				Plugin: Plugin{
					Schedule: "@daily",
					Type:     "pkginfo",
				},
				StrFunc: pkginfo.String,
			},
		}
		if envinfo.Env.CloudProvider == "aws" {
			awsPlugins := map[string]BuiltinPlugin{
				"ebsinfo": {
					Plugin: Plugin{
						Schedule: "@daily",
						Type:     "ebsinfo",
					},
					StrFunc: ebsinfo.String,
				},
			}
			for k, v := range awsPlugins {
				builtinPlugins[k] = v
			}
		}
		handleOverrides(builtinPlugins)
		for k, v := range builtinPlugins {
			v.Label = k
			builtinPlugins[k] = v
		}
	}
	return builtinPlugins
}

func getCustomPlugins() map[string]CustomPlugin {
	if len(customPlugins) == 0 {
		err := viper.UnmarshalKey("plugins.custom", &customPlugins)
		if err != nil {
			util.CustomOut(nil, err)
		}
		for k, v := range customPlugins {
			v.Label = k
			customPlugins[k] = v
		}
	}
	return customPlugins
}

// getPutURL will return a presigned url msg or error
func getPutURL(asset *Asset) (getURLMsg, error) {
	if err := asset.SubscribeURLTopic(); err != nil {
		return getURLMsg{}, err
	}
	path := fmt.Sprintf("mirach/url/put/%s/%s", asset.cust.id, asset.id)
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

func handleOverrides(builtins map[string]BuiltinPlugin) {
	overrides := map[string]BuiltinPlugin{}
	err := viper.UnmarshalKey("plugins.builtin", &overrides)
	if err != nil {
		jww.ERROR.Println(err)
	}
	for label, override := range overrides {
		if builtin, in := builtins[label]; in {
			if override.Disabled {
				builtin.Disabled = override.Disabled
			}
			if override.LoadDelay != "" {
				builtin.LoadDelay = override.LoadDelay
			}
			if override.Schedule != "" {
				builtin.Schedule = override.Schedule
			}
			builtins[label] = builtin
		}
	}
}

func handlePlugins(asset *Asset, cron *cron.MirachCron) {
	cron.Start()
	loadBuiltinPlugins(asset, cron)
	loadCustomPlugins(asset, cron)
}

func loadBuiltinPlugins(asset *Asset, cron *cron.MirachCron) {
	for _, p := range getBuiltinPlugins() {
		p.loadPlugin(cron, p.Run(asset))
	}
}

func loadCustomPlugins(asset *Asset, cron *cron.MirachCron) {
	for _, c := range getCustomPlugins() {
		// Loop over internal plugins to check name collisions.
		ok := true
		for _, b := range getBuiltinPlugins() {
			if c.Label == b.Label || c.Type == b.Type {
				err := fmt.Errorf("refusing to load plugin %v: conflicts with built-in", c.Label)
				util.CustomOut(nil, err)
				ok = false
				break
			}
		}
		if !ok {
			continue
		}
		c.loadPlugin(cron, c.Run(asset))
	}
}

// PutData Gets presigned url and put data to it. Return string of url it has
// been put to
func PutData(b []byte, asset *Asset) (string, error) {
	client := &http.Client{}
	presigned, err := getPutURL(asset)
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

// SendChunks splits a byte slice and return the number of chunks, ID of
// the chunk group, hex digest of the full byte slice's md5 hash, or any
// errors generated along the way.
func SendChunks(b []byte, asset *Asset) (int, string, error) {
	id := fmt.Sprintf("%s", uuid.New())
	var n int
	splits, err := util.SplitAt(b, MaxMQTTDataSize)
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
	case len(b) > MaxChunkedSize:
		url, err := PutData(b, asset)
		m := putHTTPMsg{msg, url}
		msgB, err = json.Marshal(m)
		if err != nil {
			return err
		}
	case len(b) > MaxMQTTDataSize:
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
