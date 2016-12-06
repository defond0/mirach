package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/viper"
)

var handler MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())
}

func Client() (c MQTT.Client) {
	asset_keys := viper.GetStringMapString("asset_keys")
	ca, err := ioutil.ReadFile(asset_keys["ca"])
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	key, err := ioutil.ReadFile(asset_keys["private_key"])
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	pair, err := tls.X509KeyPair(ca, key)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(ca)
	conf := &tls.Config{
		Certificates:       []tls.Certificate{pair},
		RootCAs:            pool,
		InsecureSkipVerify: true,
	}
	conf.BuildNameToCertificate()
	//create a ClientOptions struct setting the broker address, clientid, turn
	//off trace output and set the default message handler
	opts := MQTT.NewClientOptions().AddBroker("***REMOVED***")
	opts.SetTLSConfig(conf)
	opts.SetClientID("mirach")
	opts.SetDefaultPublishHandler(handler)
	//create and start a client using the above ClientOptions
	c = MQTT.NewClient(opts)
	token := c.Connect()
	if token.Wait() && token.Error() != nil {
		fmt.Println(err)
		panic(token.Error())
	}
	//subscribe to the topic mirach/data and request messages to be delivered
	//at a maximum qos of zero, wait for the receipt to confirm the subscription
	custID := viper.GetString("customer_id")
	assetID := viper.GetString("asset_id")
	path := fmt.Sprintf("mirach/data/%s/%s", custID, assetID)
	if token := c.Subscribe(path, 0, nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		panic(token.Error())
	}
	return c
}

func Publish(mes string, path string, c MQTT.Client) {
	token := c.Publish(path, 0, false, mes)
	token.Wait()
}
