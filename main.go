// Mirach is a tool to get information about a machine and send it to a central repository.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"

	// may use v2 so we can remove the jobs
	// "gopkg.in/robfig/cron.v2"

	"github.com/golang/glog"
	"github.com/robfig/cron"
	"github.com/spf13/viper"
)

type Plugin struct {
	Label    string `json:"label"`
	Cmd      string `json:"cmd"`
	Schedule string `json:"schedule"`
}

type result struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

func getConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/mirach/")
	viper.AddConfigPath("$HOME/.config/mirach")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	viper.WatchConfig()
}

func RunPlugin(p Plugin) func() {
	return func() {
		glog.Infof("Running plugin: %s", p.Cmd)
		cmd := exec.Command(p.Cmd)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Fatal(err)
		}
		if err := cmd.Start(); err != nil {
			log.Fatal(err)
		}
		var res result
		if err := json.NewDecoder(stdout).Decode(&res); err != nil {
			log.Fatal(err)
		}
		err = cmd.Wait()
		fmt.Printf("type: %s, data: %s\n", res.Type, res.Data)
	}
}

func main() {
	flag.Parse()
	getConfig()
	c := cron.New()
	c.Start()
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)
	plugins := make(map[string]Plugin)
	err := viper.UnmarshalKey("plugins", &plugins)
	if err != nil {
		log.Fatal(err)
	}
	for k, v := range plugins {
		glog.Infof("Adding to plugin: %s", k)
		c.AddFunc(v.Schedule, RunPlugin(v))
	}
	for _ = range s {
		// sig is a ^c, handle it
		glog.Infof("SIGINT, stopping")
		c.Stop()
		os.Exit(1)
	}
}
