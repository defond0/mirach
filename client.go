package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

// NewClient creates and connects to a new MQTT client.
func NewClient(ca []byte, key []byte, id string) (c MQTT.Client) {
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
	opts := MQTT.NewClientOptions().AddBroker("***REMOVED***")
	opts.SetTLSConfig(conf)
	opts.SetClientID(id)
	c = MQTT.NewClient(opts)
	token := c.Connect()
	if token.Wait() && token.Error() != nil {
		fmt.Println(err)
		panic(token.Error())
	}
	return c
}
