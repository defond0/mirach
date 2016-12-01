package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

var handler MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())
}

func main() {
	cert, err := ioutil.ReadFile("./6525356d25-certificate.pem.crt")
	if err != nil {
		panic(err)
	}
	key, err := ioutil.ReadFile("./6525356d25-private.pem.key")
	if err != nil {
		panic(err)
	}
	pair, err := tls.X509KeyPair(cert, key)
	if err != nil {
		panic(err)
	}

	ca, err := ioutil.ReadFile("./ca.pem")
	if err != nil {
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
	// opts.SetCleanSession(true)
	opts.SetClientID("notmirach")
	opts.SetDefaultPublishHandler(handler)

	//create and start a client using the above ClientOptions
	c := MQTT.NewClient(opts)
	token := c.Connect()

	if token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	//subscribe to the topic /go-mqtt/sample and request messages to be delivered
	//at a maximum qos of zero, wait for the receipt to confirm the subscription
	if token := c.Subscribe("mirach/data", 0, nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	//Publish 5 messages to /go-mqtt/sample at qos 1 and wait for the receipt
	//from the server after sending each message
	for i := 0; i < 5; i++ {
		text := fmt.Sprintf("this is msg #%d!", i)
		token := c.Publish("mirach/data", 0, false, text)
		token.Wait()
	}

	time.Sleep(3 * time.Second)

	//unsubscribe from /go-mqtt/sample
	if token := c.Unsubscribe("mirach/data"); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	c.Disconnect(250)
}
