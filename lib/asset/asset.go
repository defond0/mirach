package asset

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/cleardataeng/mirach/lib/cust"
	"github.com/cleardataeng/mirach/lib/util"
	"github.com/coreos/etcd/client"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/theherk/viper"
)

// Asset is a Mirach IoT thing representing this machine.
type Asset struct {
	client.Node

	// CmdChan receives command messages.
	CmdChan chan CmdMsg

	// CmdHandler is the received message handler for commands.
	CmdHandler mqtt.MessageHandler

	// Cust is the customer IoT node.
	Cust *cust.Node

	// URLChan receives URL messages.
	URLChan chan getURLMsg

	// URLHandler is the received message handler for requested URL.
	URLHandler mqtt.MessageHandler
}

// CmdMsg is a json response object from IoT containing an asset command.
type CmdMsg struct {
	Cmd string `json:"cmd"`
}

func getCustomer() (*cust.Node, error) {
	cust := new(cust.Node)
	err := cust.Init()
	if err != nil {
		msg := "customer initialization failed"
		util.CustomOut(msg, err)
		return nil, err
	}
	return cust, nil
}

// CheckRegistration returns a bool indicating if the asset seems registered.
// To be registered an asset key file, and customerID in the configuration are required.
func (n *Node) CheckRegistration(c *Customer) bool {
	var err error
	if loc := viper.GetString("asset.keys.private_key_path"); loc != "" {
		n.privKeyPath = loc
	} else {
		n.privKeyPath, err = util.FindInDirs(filepath.Join("asset", "keys", "private.pem.key"), confDirs)
		if err != nil {
			return false
		}
	}
	if exists, _ := util.Exists(n.privKeyPath); !exists {
		return false
	}
	jww.DEBUG.Println("asset already registered when checked")
	return true
}

// Init initializes an Asset MirachNode.
func (n *Node) Init() error {
	n.URLChan = make(chan getURLMsg, 1)
	var err error
	n.cust, err = getCustomer()
	if err != nil {
		return err
	}
	n.id = viper.GetString("asset.id")
	if n.id == "" {
		n.id = readAssetID()
		viper.Set("asset.id", n.id)
		err = viper.WriteConfig()
		if err != nil {
			return err
		}
	}
	if !n.CheckRegistration(n.cust) {
		err := n.Register(n.cust)
		if err != nil {
			return err
		}
	}
	if loc := viper.GetString("asset.keys.private_key_path"); loc != "" {
		n.privKeyPath = loc
	} else {
		n.privKeyPath, err = util.FindInDirs(filepath.Join("asset", "keys", "private.pem.key"), confDirs)
		if err != nil {
			return errors.New("asset private key not found")
		}
	}
	n.privKey, err = util.ReadFile(n.privKeyPath)
	if err != nil {
		return err
	}
	if loc := viper.GetString("asset.keys.cert_path"); loc != "" {
		n.certPath = loc
	} else {
		n.certPath, err = util.FindInDirs(filepath.Join("asset", "keys", "cn.pem.crt"), confDirs)
		if err != nil {
			return errors.New("asset cert not found")
		}
	}
	n.cert, err = util.ReadFile(n.certPath)
	if err != nil {
		return err
	}
	ca, err := util.GetCA(confDirs)
	if err != nil {
		return err
	}
	n.client, err = NewClient(viper.GetString("broker"), ca, n.privKey, n.cert, n.cust.id+":"+n.id)
	if err != nil {
		return errors.New("asset client connection failed")
	}
	n.cmdChan = make(chan CmdMsg, 1)
	n.cmdHandler = func(client mqtt.Client, msg mqtt.Message) {
		res := CmdMsg{}
		err := json.Unmarshal(msg.Payload(), &res)
		if err != nil {
			panic(err)
		}
		n.cmdChan <- res
	}
	path := fmt.Sprintf("mirach/cmd/%s/%s", n.cust.id, n.id)
	if subToken := n.client.Subscribe(path, 1, n.cmdHandler); subToken.Wait() && subToken.Error() != nil {
		panic(subToken.Error())
	}
	return nil
}

func (n *Node) readCmds() error {
	go func() {
		for {
			msg := <-n.cmdChan
			util.CustomOut("cmd received: "+msg.Cmd, nil)
		}
	}()
	util.CustomOut("command channel open", nil)
	return nil
}

// Register an IoT asset using a customer's client cert.
func (n *Node) Register(c *Customer) error {
	ca, err := util.GetCA()
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
	path := fmt.Sprintf("mirach/register/%s/%s", c.id, n.id)
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
		keyPath := filepath.Join(util.GetConfDirs()["sys"], "asset", "keys")
		if err := util.ForceWrite(filepath.Join(keyPath, "cn.pem.crt"), res.Cert); err != nil {
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

// SubscribeURLTopic is a function to new up a subscription to s3/put/url topic
func (n *Node) SubscribeURLTopic() error {
	custID := viper.GetString("customer.id")
	assetID := viper.GetString("asset.id")
	path := fmt.Sprintf("mirach/url/put/%s/%s", custID, assetID)
	urlHandler := func(c mqtt.Client, msg mqtt.Message) {
		res := getURLMsg{}
		empty := getURLMsg{}
		_ = json.Unmarshal(msg.Payload(), &res)
		if res != empty {
			n.urlChan <- res
		}
	}
	if subToken := n.client.Subscribe(path, 1, urlHandler); subToken.Wait() && subToken.Error() != nil {
		return subToken.Error()
	}
	return nil
}
