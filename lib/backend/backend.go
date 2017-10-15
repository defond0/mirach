// Package backend provides resources for data messaging and storage.
// We implement basic io interfaces and use these for storing data and
// for sending and receiving data from cloud services.
package backend

import mqtt "github.com/eclipse/paho.mqtt.golang"

// HTTP is for sending and receiving messages via HTTP.
type HTTP struct{}

// MQTT is for sending and receiving messages via MQTT.
type MQTT struct {
	Broker string
	Client mqtt.Client
}

// SQLite is for local storage.
type SQLite struct{}

// NewMQTT authenticates and returns a new MQTT backend.
func NewMQTT(privKey, cert []byte) *MQTT {
	return &MQTT{}
}
