// Mirach is a tool to get information about a machine and send it to a central repository.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"

	// may use v2 so we can remove the jobs
	// "gopkg.in/robfig/cron.v2"

	flags "github.com/jessevdk/go-flags"
	"github.com/robfig/cron"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

var sysConfDir string
var userConfDir string
var configDirs []string
var timeout = make(chan bool, 1)

var opts struct {
	// Slice of bool will append 'true' each time the option
	// is encountered (can be set multiple times, like -vvv)
	Verbose []bool `short:"v" long:"verbose" description:"Show verbose debug information"`
}

func getConfig() string {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	for _, d := range configDirs {
		viper.AddConfigPath(d)
	}
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	viper.SetEnvPrefix("mirach")
	viper.AutomaticEnv()
	viper.WatchConfig()
	return viper.ConfigFileUsed()
}

func main() {
	// flag.Parse()
	_, err := flags.Parse(&opts)
	if err != nil {
		panic(err)
	}
	v := len(opts.Verbose)
	switch {
	case v == 1:
		jww.SetStdoutThreshold(jww.LevelInfo)
	case v > 1:
		jww.SetStdoutThreshold(jww.LevelTrace)
	}
	if runtime.GOOS == "windows" {
		userConfDir = "%APPDATA%\\mirach"
		sysConfDir = "%PROGRAMDATA%\\mirach"
	} else {
		userConfDir = "$HOME/.config/mirach"
		sysConfDir = "/etc/mirach/"
	}
	configDirs = append(configDirs, ".", userConfDir, sysConfDir)
	getConfig()
	assetID := viper.GetString("asset_id")
	if assetID == "" {
		assetID = readAssetID()
	}
	viper.Set("asset_id", assetID)
	err = viper.WriteConfig()
	if err != nil {
		panic(err)
	}

	cust := new(Customer)
	err = cust.Init()
	if err != nil {
		jww.ERROR.Println("customer initialization failed")
	}
	asset := new(Asset)
	err = asset.Init(cust)
	if err != nil {
		jww.ERROR.Println("asset initialization failed")
	}

	plugins := make(map[string]Plugin)
	err = viper.UnmarshalKey("plugins", &plugins)
	if err != nil {
		jww.ERROR.Println(err)
	}
	cron := cron.New()
	cron.Start()
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)
	for k, v := range plugins {
		jww.INFO.Printf("Adding to plugin: %s", k)
		cron.AddFunc(v.Schedule, RunPlugin(v, asset.client))
	}
	for _ = range s {
		// sig is a ^c, handle it
		jww.INFO.Println("SIGINT, stopping")
		cron.Stop()
		os.Exit(1)
	}
}
