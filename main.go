// Mirach is a tool to get information about a machine and send it to a central repository.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
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
var verbosity int

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
	return viper.ConfigFileUsed()
}

func main() {
	_, err := flags.Parse(&opts)
	if err != nil {
		panic(err)
	}
	verbosity = len(opts.Verbose)
	switch {
	case verbosity == 1:
		jww.SetStdoutThreshold(jww.LevelInfo)
	case verbosity > 1:
		jww.SetStdoutThreshold(jww.LevelTrace)
	}
	if runtime.GOOS == "windows" {
		userConfDir = filepath.Join("%APPDATA%", "mirach")
		sysConfDir = filepath.Join("%PROGRAMDATA%", "mirach")
	} else {
		userConfDir = "$HOME/.config/mirach"
		sysConfDir = "/etc/mirach/"
	}
	configDirs = append(configDirs, ".", userConfDir, sysConfDir)
	getConfig()

	cust := new(Customer)
	err = cust.Init()
	if err != nil {
		msg := "customer initialization failed"
		customOut(msg, err)
		os.Exit(1)
	}
	asset := new(Asset)
	err = asset.Init(cust)
	if err != nil {
		msg := "asset initialization failed"
		customOut(msg, err)
		os.Exit(1)
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
	err = asset.readCmds()
	if err != nil {
		msg := "stopped receiving commands; stopping mirach"
		customOut(msg, err)
		os.Exit(1)
	}
	msg := "mirach entered running state; plugins loaded"
	customOut(msg, nil)
	for _ = range s {
		// sig is a ^c, handle it
		jww.DEBUG.Println("SIGINT, stopping")
		cron.Stop()
		os.Exit(1)
	}
}
