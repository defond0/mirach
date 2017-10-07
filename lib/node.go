package lib

import mqtt "github.com/eclipse/paho.mqtt.golang"

// MirachNode is an IoT thing.
type MirachNode struct {
	id          string
	privKeyPath string
	privKey     []byte
	certPath    string
	cert        []byte
	client      mqtt.Client
	regHandler  mqtt.MessageHandler
	regMsg      chan RegMsg // channel receiving registration messages
}

// RegMsg is a json response object from IoT containing client cert keys.
type RegMsg struct {
	PrivKey string `json:"private_key"`
	PubKey  string `json:"public_key"`
	Cert    string `json:"certificate"`
	CA      string `json:"ca"`
}
