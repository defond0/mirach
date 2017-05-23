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

	cust       *Customer
	cmdHandler mqtt.MessageHandler
	urlHandler mqtt.MessageHandler
	cmdChan    chan CmdMsg    // channel receiving command messages
	urlChan    chan getURLMsg // channel receiving url messages
}

func getCustomer() (*Customer, error) {
	cust := new(Customer)
	err := cust.Init()
	if err != nil {
		msg := "customer initialization failed"
		util.CustomOut(msg, err)
		return nil, err
	}
	return cust, nil
}

// Init initializes an Asset MirachNode.
func (a *Asset) Init() error {
	a.urlChan = make(chan getURLMsg, 1)
	var err error
	a.cust, err = getCustomer()
	if err != nil {
		return err
	}
	a.id = viper.GetString("asset.id")
	if a.id == "" {
		a.id = readAssetID()
		viper.Set("asset.id", a.id)
		err = viper.WriteConfig()
		if err != nil {
			return err
		}
	}
	if !a.CheckRegistration(a.cust) {
		err := a.Register(a.cust)
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
	a.client, err = NewClient(viper.GetString("broker"), ca, a.privKey, a.cert, a.cust.id+":"+a.id)
	if err != nil {
		return errors.New("asset client connection failed")
	}
	a.cmdChan = make(chan CmdMsg, 1)
	a.cmdHandler = func(client mqtt.Client, msg mqtt.Message) {
		res := CmdMsg{}
		err := json.Unmarshal(msg.Payload(), &res)
		if err != nil {
			panic(err)
		}
		a.cmdChan <- res
	}
	path := fmt.Sprintf("mirach/cmd/%s/%s", a.cust.id, a.id)
	if subToken := a.client.Subscribe(path, 1, a.cmdHandler); subToken.Wait() && subToken.Error() != nil {
		panic(subToken.Error())
	}
	return nil
}

// SubscribeURLTopic is a function to new up a subscription to s3/put/url topic
func (a *Asset) SubscribeURLTopic() error {
	custID := viper.GetString("customer.id")
	assetID := viper.GetString("asset.id")
	path := fmt.Sprintf("mirach/url/put/%s/%s", custID, assetID)
	urlHandler := func(c mqtt.Client, msg mqtt.Message) {
		res := getURLMsg{}
		empty := getURLMsg{}
		_ = json.Unmarshal(msg.Payload(), &res)
		if res != empty {
			a.urlChan <- res
		}
	}
	if subToken := a.client.Subscribe(path, 1, urlHandler); subToken.Wait() && subToken.Error() != nil {
		return subToken.Error()
	}
	return nil
}

// Register an IoT asset using a customer's client cert.
func (a *Asset) Register(c *Customer) error {
	ca, err := util.GetCA(confDirs)
	if err != nil {
		return err
	}
	c.client, err = NewClient(viper.GetString("broker"), ca, c.privKey, c.cert, c.id)
	defer c.client.Disconnect(0)
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
	timeoutCh := util.Timeout(15 * time.Second)
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
			msg := <-a.cmdChan
			util.CustomOut("cmd received: "+msg.Cmd, nil)
		}
	}()
	util.CustomOut("command channel open", nil)
	return nil
}
