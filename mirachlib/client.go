package mirachlib

import (
	"crypto/tls"
	"crypto/x509"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// NewClient creates and connects to a new MQTT client.
func NewClient(broker string, ca, privKey, cert []byte, id string) (mqtt.Client, error) {
	pair, err := tls.X509KeyPair(cert, privKey)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(ca)
	conf := &tls.Config{
		Certificates:       []tls.Certificate{pair},
		RootCAs:            pool,
		InsecureSkipVerify: true,
	}
	conf.BuildNameToCertificate()
	options := mqtt.NewClientOptions().AddBroker(broker)
	options.SetTLSConfig(conf)
	options.SetClientID(id)
	c := mqtt.NewClient(options)
	token := c.Connect()
	if token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return c, nil
}

// PubWait publishes a byte slice to a given path using a given client.
func PubWait(c mqtt.Client, path string, b []byte) error {
	token := c.Publish(path, 1, false, b)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}
