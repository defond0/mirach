package mirachlib

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"gitlab.eng.cleardata.com/dash/mirach/util"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

// Customer is a Mirach IoT thing representing this customer.
type Customer struct {
	MirachNode
	idHandler mqtt.MessageHandler
	idMsg     chan CustIDMsg // channel receiving customer ID messages
}

// CustIDMsg is a json response object from IoT containing customer ID.
type CustIDMsg struct {
	ID string `json:"customer_id"`
}

// Init initializes a Customer MirachNode.
func (c *Customer) Init() error {
	var err error
	if loc := viper.GetString("customer.keys.private_key_path"); loc != "" {
		c.privKeyPath = loc
	} else {
		c.privKeyPath, err = util.FindInDirs(filepath.Join("customer", "keys", "private.pem.key"), confDirs)
		if err != nil {
			return errors.New("customer private key not found")
		}
	}
	c.privKey, err = util.ReadFile(c.privKeyPath)
	if err != nil {
		return err
	}
	if loc := viper.GetString("customer.keys.cert_path"); loc != "" {
		c.certPath = loc
	} else {
		c.certPath, err = util.FindInDirs(filepath.Join("customer", "keys", "ca.pem.crt"), confDirs)
		if err != nil {
			return errors.New("customer cert not found")
		}
	}
	c.cert, err = util.ReadFile(c.certPath)
	if err != nil {
		return err
	}
	ca, err := util.GetCA(confDirs)
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

// GetCustomerID returns the customer id or an error.
func (c *Customer) GetCustomerID() (string, error) {
	c.id = viper.GetString("customer.id")
	if c.id != "" {
		return c.id, nil
	}
	if c.privKey == nil {
		c.Init()
	}
	ca, err := util.GetCA(confDirs)
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
	timeoutCh := util.Timeout(10 * time.Second)
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
