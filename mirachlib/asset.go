package mirachlib

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"gitlab.eng.cleardata.com/dash/mirach/util"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/theherk/viper"
)

// CmdMsg is a json response object from IoT containing an asset command.
type CmdMsg struct {
	Cmd string `json:"cmd"`
}

// Asset is a Mirach IoT thing representing this machine.
type Asset struct {
	MirachNode
	cmdHandler mqtt.MessageHandler
	cmdMsg     chan CmdMsg // channel receiving command messages
}

// Init initializes an Asset MirachNode.
func (a *Asset) Init(c *Customer) error {
	var err error
	a.id = viper.GetString("asset.id")
	if a.id == "" {
		a.id = readAssetID()
		viper.Set("asset.id", a.id)
		err = viper.WriteConfig()
		if err != nil {
			panic(err)
		}
	}
	if !a.CheckRegistration(c) {
		err := a.Register(c)
		if err != nil {
			return err
		}
	}
	if loc := viper.GetString("asset.keys.private_key_path"); loc != "" {
		a.privKeyPath = loc
	} else {
		a.privKeyPath, err = util.FindInDirs(filepath.Join("asset", "keys", "private.pem.key"), confDirs)
		if err != nil {
			return errors.New("asset private key not found")
		}
	}
	a.privKey, err = util.ReadFile(a.privKeyPath)
	if err != nil {
		return err
	}
	if loc := viper.GetString("asset.keys.cert_path"); loc != "" {
		a.certPath = loc
	} else {
		a.certPath, err = util.FindInDirs(filepath.Join("asset", "keys", "ca.pem.crt"), confDirs)
		if err != nil {
			return errors.New("asset cert not found")
		}
	}
	a.cert, err = util.ReadFile(a.certPath)
	if err != nil {
		return err
	}
	ca, err := util.GetCA(confDirs)
	if err != nil {
		return err
	}
	a.client, err = NewClient(ca, a.privKey, a.cert, c.id+":"+a.id)
	if err != nil {
		return errors.New("asset client connection failed")
	}
	a.cmdMsg = make(chan CmdMsg, 1)
	a.cmdHandler = func(client mqtt.Client, msg mqtt.Message) {
		res := CmdMsg{}
		err := json.Unmarshal(msg.Payload(), &res)
		if err != nil {
			panic(err)
		}
		a.cmdMsg <- res
	}
	path := fmt.Sprintf("mirach/cmd/%s/%s", c.id, a.id)
	if subToken := a.client.Subscribe(path, 1, a.cmdHandler); subToken.Wait() && subToken.Error() != nil {
		panic(subToken.Error())
	}
	return nil
}

// Register an IoT asset using a customer's client cert.
func (a *Asset) Register(c *Customer) error {
	c.regMsg = make(chan RegMsg, 1)
	c.regHandler = func(client mqtt.Client, msg mqtt.Message) {
		res := RegMsg{}
		err := json.Unmarshal(msg.Payload(), &res)
		if err != nil {
			panic(err)
		}
		c.regMsg <- res
	}
	path := fmt.Sprintf("mirach/register/%s/%s", c.id, a.id)
	pubToken := c.client.Publish(path, 1, false, "")
	if !pubToken.WaitTimeout(15 * time.Second) {
		return errors.New("failed while registering; check credentials/config")
	}
	if subToken := c.client.Subscribe(path, 1, c.regHandler); subToken.Wait() && subToken.Error() != nil {
		return subToken.Error()
	}
	timeoutCh := util.Timeout(10 * time.Second)
	select {
	case res := <-c.regMsg:
		keyPath := filepath.Join(sysConfDir, "asset", "keys")
		if err := util.ForceWrite(filepath.Join(keyPath, "ca.pem.crt"), res.Cert); err != nil {
			return err
		}
		if err := util.ForceWrite(filepath.Join(keyPath, "public.pem.key"), res.PubKey); err != nil {
			return err
		}
		if err := util.ForceWrite(filepath.Join(keyPath, "private.pem.key"), res.PrivKey); err != nil {
			return err
		}
	case <-timeoutCh:
		return errors.New("failed while registering; check credentials")
	}
	return nil
}

// CheckRegistration returns a bool indicating if the asset seems registered.
// To be registered an asset key file, and customerID in the configuration are required.
func (a *Asset) CheckRegistration(c *Customer) bool {
	var err error
	if loc := viper.GetString("asset.keys.private_key_path"); loc != "" {
		a.privKeyPath = loc
	} else {
		a.privKeyPath, err = util.FindInDirs(filepath.Join("asset", "keys", "private.pem.key"), confDirs)
		if err != nil {
			return false
		}
	}
	if exists, _ := util.Exists(a.privKeyPath); !exists {
		return false
	}
	jww.DEBUG.Println("asset already registered when checked")
	return true
}

func (a *Asset) readCmds() error {
	go func() {
		for {
			msg := <-a.cmdMsg
			CustomOut("cmd received: "+msg.Cmd, nil)
		}
	}()
	CustomOut("command channel open", nil)
	return nil
}
