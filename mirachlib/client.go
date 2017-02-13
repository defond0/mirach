package mirach

import (
	"crypto/tls"
	"crypto/x509"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// NewClient creates and connects to a new MQTT client.
func NewClient(ca, privKey, cert []byte, id string) (mqtt.Client, error) {
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
	options := mqtt.NewClientOptions().AddBroker("***REMOVED***")
	options.SetTLSConfig(conf)
	options.SetClientID(id)
	c := mqtt.NewClient(options)
	token := c.Connect()
	if token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return c, nil
}
