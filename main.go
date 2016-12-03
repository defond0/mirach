// Mirach is a tool to get information about a machine and send it to a central repository.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	// may use v2 so we can remove the jobs
	// "gopkg.in/robfig/cron.v2"

	"github.com/golang/glog"
	"github.com/robfig/cron"
	"github.com/spf13/viper"
)

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
	viper.SetEnvPrefix("mirach")
	viper.AutomaticEnv()
	viper.WatchConfig()
}

func main() {
	flag.Parse()
	flag.Lookup("logtostderr").Value.Set("true")
	getConfig()
	AssetID := viper.GetString("asset_id")
	if AssetID == "" {
		AssetID = readAssetID()
	}
	fmt.Println(AssetID)
	plugins := make(map[string]Plugin)
	err := viper.UnmarshalKey("plugins", &plugins)
	if err != nil {
		log.Fatal(err)
	}
	c := cron.New()
	c.Start()
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)
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
