package cust

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/cleardataeng/mirach/lib/client"
	"github.com/cleardataeng/mirach/lib/util"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/theherk/viper"
)

// CustIDMsg is a json response object from IoT containing customer ID.
type CustIDMsg struct {
	ID string `json:"customer_id"`
}

// Customer is a Mirach IoT thing representing this customer.
type Node struct {
	client.Node

	// IDHandler is the received message handler for customer ID messages.
	IDHandler mqtt.MessageHandler

	// IDMsg receives customer ID messages.
	IDMsg chan CustIDMsg
}

// GetCustomerID returns the customer id or an error.
func (n *Node) GetCustomerID() (string, error) {
	n.ID = viper.GetString("customer.id")
	if n.ID != "" {
		return n.ID, nil
	}
	if n.PrivKey == nil {
		n.Init()
	}
	ca, err := util.GetCA()
	client, err := client.NewClient(viper.GetString("broker"), ca, n.PrivKey, n.Cert, "mirach-registration-client")
	if err != nil {
		return "", errors.New("registration client connection failed")
	}
	defer client.Disconnect(0)
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
	timeoutCh := util.Timeout(10 * time.Second)
	select {
	case res := <-idMsg:
		n.ID = res.ID
	case <-timeoutCh:
		return n.ID, errors.New("failed while getting customer_id; check credentials")
	}
	viper.Set("customer.id", n.ID)
	err = viper.WriteConfig()
	if err != nil {
		panic(err)
	}
	return n.ID, nil
}

// Init initializes a Customer MirachNode.
func (n *Node) Init() error {
	var err error
	if loc := viper.GetString("customer.keys.private_key_path"); loc != "" {
		n.PrivKeyPath = loc
	} else {
		n.PrivKeyPath, err = util.FindInConfDirs(filepath.Join("customer", "keys", "private.pem.key"))
		if err != nil {
			return errors.New("customer private key not found")
		}
	}
	n.PrivKey, err = util.ReadFile(n.PrivKeyPath)
	if err != nil {
		return err
	}
	if loc := viper.GetString("customer.keys.cert_path"); loc != "" {
		n.CertPath = loc
	} else {
		n.CertPath, err = util.FindInConfDirs(filepath.Join("customer", "keys", "ca.pem.crt"))
		if err != nil {
			return errors.New("customer cert not found")
		}
	}
	n.Cert, err = util.ReadFile(n.CertPath)
	if err != nil {
		return err
	}
	if _, err = n.GetCustomerID(); err != nil {
		return err
	}
	if err != nil {
		return errors.New("customer client connection failed")
	}
	return nil
}
