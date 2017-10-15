package client

import mqtt "github.com/eclipse/paho.mqtt.golang"

// Node is an IoT thing.
type Node struct {
	ID          string
	PrivKeyPath string
	PrivKey     []byte
	CertPath    string
	Cert        []byte
	Client      mqtt.Client
	RegHandler  mqtt.MessageHandler
	RegMsg      chan RegMsg // channel receiving registration messages
}

// RegMsg is a json response object from IoT containing client cert keys.
type RegMsg struct {
	PrivKey string `json:"private_key"`
	PubKey  string `json:"public_key"`
	Cert    string `json:"certificate"`
	CA      string `json:"ca"`
}
