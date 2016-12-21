// Mirach is a tool to get information about a machine and send it to a central repository.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"time"

	// may use v2 so we can remove the jobs
	// "gopkg.in/robfig/cron.v2"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/golang/glog"
	"github.com/google/uuid"
	"github.com/robfig/cron"
	"github.com/spf13/viper"
)

// Asset is a Mirach IoT thing representing this machine.
type Asset struct {
	id         string
	privKeyLoc string
	privKey    []byte
	client     mqtt.Client
}

// Customer is a Mirach IoT thing representing this customer.
type Customer struct {
	id         string
	privKeyLoc string
	privKey    []byte
	client     mqtt.Client
	resHandler mqtt.MessageHandler
	res        chan CustResponse // channel receiving responses
}

// CustResponse is a json response object from IoT.
type CustResponse struct {
	ID      string `json:"customer_id"`
	PrivKey string `json:"private_key"`
	CA      string `json:"ca"`
}

var ca []byte
var timeout = make(chan bool, 1)

func getConfig() string {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	if runtime.GOOS == "windows" {
		viper.AddConfigPath("%PROGRAMDATA%\\mirach")
		viper.AddConfigPath("%APPDATA%\\mirach")
	} else {
		viper.AddConfigPath("/etc/mirach/")
		viper.AddConfigPath("$HOME/.config/mirach")
	}
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	viper.SetEnvPrefix("mirach")
	viper.AutomaticEnv()
	viper.WatchConfig()
	return viper.ConfigFileUsed()
}

// GetCustomerID returns the customer id or an error.
func (c *Customer) GetCustomerID() (string, error) {
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
		ca := res.CA
		// TODO: Write the CA to a file using runtime.GOOS constants
	case <-timeout:
		return errors.New("failed while registering; check credentials")
	}
	return nil
}

// CheckRegistration returns a bool indicating if the asset seems registered.
// To be registered an asset key file, and customerID in the configuration are required.
func (a *Asset) CheckRegistration(c *Customer) bool {
	if _, err := os.Stat(a.privKeyLoc); os.IsNotExist(err) {
		return false
	}
	if c.id == "" {
		return false
	}
	return true
}

func main() {
	flag.Parse()
	err := flag.Lookup("logtostderr").Value.Set("true")
	if err != nil {
		glog.Infof("unable to log to stderr")
	}
	configFile := getConfig()
	configPath := filepath.Dir(configFile)
	assetID := viper.GetString("asset_id")
	if assetID == "" {
		assetID = readAssetID()
	}
	viper.Set("asset_id", assetID)
	err = viper.WriteConfig()
	if err != nil {
		panic(err)
	}

	cust := new(Customer)
	asset := Asset{id: assetID}
	if !asset.CheckRegistration(cust) {
		asset.Register(cust)
	}
	// TODO: Get actual ca from file system
	ca, err := ioutil.ReadFile(string(ca))
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	cust.client = NewClient(cust.privKey, ca, "mirach-customer-client")
	cust.res = make(chan CustResponse, 1)
	cust.resHandler = func(client mqtt.Client, msg mqtt.Message) {
		res := CustResponse{}
		err := json.Unmarshal([]byte(msg.Payload()), &res)
		if err != nil {
			panic(err)
		}
		cust.res <- res
	}

	path := fmt.Sprintf("mirach/data/%s/%s", cust.id, assetID)
	if token := c.Subscribe(path, 0, nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		panic(token.Error())
	}

	assetClient := NewClient(assetKey, ca, assetID)

	plugins := make(map[string]Plugin)
	err = viper.UnmarshalKey("plugins", &plugins)
	if err != nil {
		log.Fatal(err)
	}
	c := cron.New()
	c.Start()
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)
	for k, v := range plugins {
		glog.Infof("Adding to plugin: %s", k)
		c.AddFunc(v.Schedule, RunPlugin(v, client))
	}
	for _ = range s {
		// sig is a ^c, handle it
		glog.Infof("SIGINT, stopping")
		c.Stop()
		os.Exit(1)
	}
}
