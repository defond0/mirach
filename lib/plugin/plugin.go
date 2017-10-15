package plugin

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

	"github.com/cleardataeng/mirach/lib/cron"
	"github.com/cleardataeng/mirach/lib/util"
	"github.com/cleardataeng/mirach/plugin/compinfo"
	"github.com/cleardataeng/mirach/plugin/ebsinfo"
	"github.com/cleardataeng/mirach/plugin/envinfo"
	"github.com/cleardataeng/mirach/plugin/pkginfo"

	"github.com/google/uuid"
	robfigCron "github.com/robfig/cron"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/theherk/viper"
)

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
	RunAtLoad bool   `mapstructure:"run_at_load"`
	Schedule  string
	Type      string
}

type putHTTPMsg struct {
	mqttMsg
	URL string `json:"url"`
}

const (
	// MaxChunkedSize is the threshold beyond which presigned urls are used.
	MaxChunkedSize = 512000

	// MaxMQTTDataSize is the max size of chunks to be sent over MQTT.
	MaxMQTTDataSize = 88000
)

var (
	builtinPlugins map[string]BuiltinPlugin
	customPlugins  map[string]CustomPlugin
)

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
					RunAtLoad: true,
					Schedule:  "@daily",
					Type:      "compinfo",
				},
				StrFunc: compinfo.GetSysString,
			},
			"pkginfo": {
				Plugin: Plugin{
					LoadDelay: "2m",
					RunAtLoad: true,
					Schedule:  "@daily",
					Type:      "pkginfo",
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
	for label, builtin := range builtins {
		path := fmt.Sprintf("plugins.builtin.%s", label)
		var override BuiltinPlugin
		meta, err := viper.UnmarshalKeyWithMeta(path, &override)
		if err != nil {
			jww.ERROR.Println(err)
		}
		if override.Disabled {
			builtin.Disabled = override.Disabled
		}
		if override.LoadDelay != "" {
			builtin.LoadDelay = override.LoadDelay
		}
		for _, k := range meta.Keys {
			if k == "run_at_load" {
				builtin.RunAtLoad = override.RunAtLoad
				break
			}
		}
		if override.Schedule != "" {
			builtin.Schedule = override.Schedule
		}
		builtins[label] = builtin
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
		jww.INFO.Printf("%s: running", pLabel)
		d := pFunc()
		if err := SendData([]byte(d), pType, asset); err != nil {
			jww.ERROR.Println(err)
		}
	}
}

// Run will run custom plugin and publishes its results.
func (p *CustomPlugin) Run(asset *Asset) func() {
	pCmd := p.Cmd
	pLabel := p.Label
	pType := p.Type
	return func() {
		jww.INFO.Printf("%s: running", pLabel)
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

func (p *Plugin) loadPlugin(cron *cron.MirachCron, f func()) {
	if p.Disabled {
		jww.INFO.Printf("%s: disabled, skipping", p.Label)
		return
	}
	var (
		addMsg     = fmt.Sprintf("%s: adding to cron", p.Label)     // notification of load process
		successMsg = fmt.Sprintf("%s: added to cron", p.Label)      // message if successfully loaded
		errorMsg   = fmt.Sprintf("%s: failed add to cron", p.Label) // error if failure to load
	)
	delay, err := time.ParseDuration(p.LoadDelay)
	if err != nil {
		if p.LoadDelay != "" {
			msg := "invalid duration: continuing without delay"
			util.CustomOut(msg, err)
			delay = 0
		}
		// If an empty sting is given an error is generated, but the
		// delay is correctly set to zero. No need to set here.
	}
	if delay > 0 {
		addMsg += fmt.Sprintf(" in %s", delay)
		successMsg += fmt.Sprintf(" after %s", delay)
		errorMsg += fmt.Sprintf(" after %s", delay)
	}
	jww.INFO.Println(addMsg)
	res := make(chan interface{})
	cron.AddFuncDelayed(p.Schedule, f, delay, res)
	go logResChan(successMsg, errorMsg, res)
	if p.RunAtLoad {
		pLabel := p.Label
		go func() {
			jww.TRACE.Printf("%s: run_at_load true; run when loaded then resume schedule", pLabel)
			time.Sleep(delay)
			f()
		}()
		return
	}
	s, _ := robfigCron.Parse(p.Schedule)
	jww.TRACE.Printf("%s: initial run time: %s", p.Label, s.Next(time.Now()))
}
