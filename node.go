package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

// MirachNode is an IoT thing.
type MirachNode struct {
	id          string
	privKeyPath string
	privKey     []byte
	certPath    string
	cert        []byte
	client      mqtt.Client
	regHandler  mqtt.MessageHandler
	regMsg      chan RegMsg // channel receiving registration messages
}

// Asset is a Mirach IoT thing representing this machine.
type Asset struct {
	MirachNode
	cmdHandler mqtt.MessageHandler
	cmdMsg     chan CmdMsg // channel receiving command messages
}

// Customer is a Mirach IoT thing representing this customer.
type Customer struct {
	MirachNode
	idHandler mqtt.MessageHandler
	idMsg     chan CustIDMsg // channel receiving customer ID messages
}

// RegMsg is a json response object from IoT containing client cert keys.
type RegMsg struct {
	PrivKey string `json:"private_key"`
	PubKey  string `json:"public_key"`
	Cert    string `json:"certificate"`
	CA      string `json:"ca"`
}

// CustIDMsg is a json response object from IoT containing customer ID.
type CustIDMsg struct {
	ID string `json:"customer_id"`
}

// CmdMsg is a json response object from IoT containing an asset command.
type CmdMsg struct {
	Cmd string `json:"cmd"`
}

// Init initializes a Customer MirachNode.
func (c *Customer) Init() error {
	var err error
	if loc := viper.GetString("customer.keys.private_key_path"); loc != "" {
		c.privKeyPath = loc
	} else {
		c.privKeyPath, err = findInDirs(filepath.Join("customer", "keys", "private.pem.key"), configDirs)
		if err != nil {
			return errors.New("customer private key not found")
		}
	}
	c.privKey, err = ioutil.ReadFile(c.privKeyPath)
	if err != nil {
		return err
	}
	if loc := viper.GetString("customer.keys.cert_path"); loc != "" {
		c.certPath = loc
	} else {
		c.certPath, err = findInDirs(filepath.Join("customer", "keys", "ca.pem.crt"), configDirs)
		if err != nil {
			return errors.New("customer cert not found")
		}
	}
	c.cert, err = ioutil.ReadFile(c.certPath)
	if err != nil {
		return err
	}
	ca, err := getCA()
	if err != nil {
		return err
	}
	if _, err = c.GetCustomerID(); err != nil {
		return err
	}
	c.client, err = NewClient(ca, c.privKey, c.cert, c.id)
	if err != nil {
		return errors.New("customer client connection failed")
	}
	return nil
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
		a.privKeyPath, err = findInDirs(filepath.Join("asset", "keys", "private.pem.key"), configDirs)
		if err != nil {
			return errors.New("asset private key not found")
		}
	}
	a.privKey, err = ioutil.ReadFile(a.privKeyPath)
	if err != nil {
		return err
	}
	if loc := viper.GetString("asset.keys.cert_path"); loc != "" {
		a.certPath = loc
	} else {
		a.certPath, err = findInDirs(filepath.Join("asset", "keys", "ca.pem.crt"), configDirs)
		if err != nil {
			return errors.New("asset cert not found")
		}
	}
	a.cert, err = ioutil.ReadFile(a.certPath)
	if err != nil {
		return err
	}
	ca, err := getCA()
	if err != nil {
		return err
	}
	a.client, err = NewClient(ca, a.privKey, a.cert, a.id)
	if err != nil {
		return errors.New("asset client connection failed")
	}
	a.cmdHandler = func(client mqtt.Client, msg mqtt.Message) {
		a.cmdMsg = make(chan CmdMsg, 1)
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

// GetCustomerID returns the customer id or an error.
func (c *Customer) GetCustomerID() (string, error) {
	c.id = viper.GetString("customer.id")
	if c.id != "" {
		return c.id, nil
	}

	if c.privKey == nil {
		c.Init()
	}
	ca, err := getCA()
	client, err := NewClient(ca, c.privKey, c.cert, "mirach-registration-client")
	if err != nil {
		return "", errors.New("registration client connection failed")
	}

	idMsg := make(chan CustIDMsg, 1)
	idHandler := func(client mqtt.Client, msg mqtt.Message) {
		res := CustIDMsg{}
		err := json.Unmarshal(msg.Payload(), &res)
		if err != nil {
			panic(err)
		}
		idMsg <- res
	}

	tempID := uuid.New()
	path := fmt.Sprintf("mirach/customer_id/%s", tempID)
	pubToken := client.Publish(path, 1, false, "")
	pubToken.Wait()
	if subToken := client.Subscribe(path, 1, idHandler); subToken.Wait() && subToken.Error() != nil {
		fmt.Println(subToken.Error())
		panic(subToken.Error())
	}
	timeoutCh := Timeout(10 * time.Second)
	select {
	case res := <-idMsg:
		c.id = res.ID
	case <-timeoutCh:
		return c.id, errors.New("failed while getting customer_id; check credentials")
	}
	viper.Set("customer.id", c.id)
	err = viper.WriteConfig()
	if err != nil {
		panic(err)
	}
	return c.id, nil
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
	pubToken.Wait()
	if subToken := c.client.Subscribe(path, 1, c.regHandler); subToken.Wait() && subToken.Error() != nil {
		panic(subToken.Error())
	}
	Timeout(30*time.Second, timeoutCh)
	select {
	case res := <-c.regMsg:
		keyPath := filepath.Join(sysConfDir, "asset", "keys")
		if err := ForceWrite(filepath.Join(keyPath, "ca.pem.crt"), res.Cert); err != nil {
			return err
		}
		if err := ForceWrite(filepath.Join(keyPath, "public.pem.key"), res.PubKey); err != nil {
			return err
		}
		if err := ForceWrite(filepath.Join(keyPath, "private.pem.key"), res.PrivKey); err != nil {
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
		a.privKeyPath, err = findInDirs(filepath.Join("asset", "keys", "private.pem.key"), configDirs)
		if err != nil {
			return false
		}
	}
	if _, err := os.Stat(a.privKeyPath); os.IsNotExist(err) {
		return false
	}
	jww.INFO.Println("asset was registered")
	return true
}
