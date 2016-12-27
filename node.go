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
	"github.com/google/uuid"
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
	resHandler  mqtt.MessageHandler
}

// Asset is a Mirach IoT thing representing this machine.
type Asset struct {
	MirachNode
	res chan AssetResponse // channel receiving responses
}

// Customer is a Mirach IoT thing representing this customer.
type Customer struct {
	MirachNode
	res chan CustResponse // channel receiving responses
}

// CustResponse is a json response object from IoT.
type CustResponse struct {
	ID      string `json:"customer_id"`
	PrivKey string `json:"private_key"`
	PubKey  string `json:"public_key"`
	Cert    string `json:"certificate"`
	CA      string `json:"ca"`
}

// AssetResponse is a json response object from IoT.
type AssetResponse struct {
	Cmd string `json:"cmd"`
}

// Init initializes a Customer MirachNode.
func (c *Customer) Init() error {
	var err error
	if loc := viper.GetString("customer.keys.private_key_path"); loc != "" {
		c.privKeyPath = loc
	} else {
		c.privKeyPath, err = findInDirs(filepath.Join("customer", "keys", "private_key.pem"), configDirs)
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
		c.certPath, err = findInDirs(filepath.Join("customer", "keys", "cert.pem"), configDirs)
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
	c.client, err = NewClient(ca, c.privKey, c.cert, "mirach-customer-client")
	if err != nil {
		return errors.New("customer client connection failed")
	}
	c.res = make(chan CustResponse, 1)
	c.resHandler = func(client mqtt.Client, msg mqtt.Message) {
		res := CustResponse{}
		err = json.Unmarshal(msg.Payload(), &res)
		if err != nil {
			panic(err)
		}
		c.res <- res
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
		a.Register(c)
	}
	if loc := viper.GetString("asset.keys.private_key_path"); loc != "" {
		a.privKeyPath = loc
	} else {
		a.privKeyPath, err = findInDirs(filepath.Join("asset", "keys", "private_key.pem"), configDirs)
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
		a.certPath, err = findInDirs(filepath.Join("asset", "keys", "cert.pem"), configDirs)
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
	a.res = make(chan AssetResponse, 1)
	a.resHandler = func(client mqtt.Client, msg mqtt.Message) {
		res := AssetResponse{}
		err := json.Unmarshal([]byte(msg.Payload()), &res)
		if err != nil {
			panic(err)
		}
		a.res <- res
	}
	path := fmt.Sprintf("mirach/cmd/%s/%s", c.id, a.id)
	if token := a.client.Subscribe(path, 0, nil); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	return nil
}

// GetCustomerID returns the customer id or an error.
func (c *Customer) GetCustomerID() (string, error) {
	c.id = viper.GetString("customer.id")
	if c.id != "" {
		return c.id, nil
	}
	tempID := uuid.New()
	path := fmt.Sprintf("mirach/customer_id/%s", tempID)
	pubToken := c.client.Publish(path, 1, false, "")
	pubToken.Wait()
	if subToken := c.client.Subscribe(path, 1, nil); subToken.Wait() && subToken.Error() != nil {
		fmt.Println(subToken.Error())
		panic(subToken.Error())
	}
	go func() {
		time.Sleep(30 * time.Second)
		timeout <- true
	}()
	select {
	case <-c.res:
		res := <-c.res
		c.id = res.ID
	case <-timeout:
		return c.id, errors.New("failed while getting customer_id; check credentials")
	}
	return c.id, nil
}

// Register an IoT asset using a customer's client cert.
func (a *Asset) Register(c *Customer) error {
	custID, err := c.GetCustomerID()
	if err != nil {
		return err
	}
	viper.Set("customer_id", custID)
	path := fmt.Sprintf("mirach/register/%s/%s", custID, a.id)
	if token := c.client.Subscribe(path, 0, nil); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	go func() {
		time.Sleep(30 * time.Second)
		timeout <- true
	}()
	select {
	case <-c.res:
		res := <-c.res
		f, err := os.Create(filepath.Join(sysConfDir, "ca.pem"))
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = f.WriteString(res.CA)
		if err != nil {
			return err
		}
	case <-timeout:
		return errors.New("failed while registering; check credentials")
	}
	return nil
}

// CheckRegistration returns a bool indicating if the asset seems registered.
// To be registered an asset key file, and customerID in the configuration are required.
func (a *Asset) CheckRegistration(c *Customer) bool {
	if _, err := os.Stat(a.privKeyPath); os.IsNotExist(err) {
		return false
	}
	if c.id == "" {
		return false
	}
	return true
}
