// Mirach is a tool to get information about a machine and send it to a central repository.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	flags "github.com/jessevdk/go-flags"
	"github.com/robfig/cron"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

//this is a mirach interface
type baseMirachSession interface {
	getConfig() string
	configureLogging()
	setConfigDirs()
	initializeConfigAndLogging()
	getCustomer() *Customer
	getAsset(cust *Customer) *Asset
	handlePlugins(client mqtt.Client, cron *cron.Cron)
	handleCommands(asset *Asset)
	getSysConfDir() string
	getUserConfDir() string
	getConfigDirs() []string
	getVerbosity() int
}

//this is a mirach struct
type mirachSession struct {
	sysConfDir  string
	userConfDir string
	configDirs  []string
	verbosity   int
	customer    *Customer
	asset       *Asset
}

var Mirach baseMirachSession

var opts struct {
	Verbose []bool `short:"v" long:"verbose" description:"Show verbose debug information"`
	Version bool   `long:"version" description:"Show version"`
}

func (s mirachSession) getVerbosity() int {
	return s.verbosity
}

func (s mirachSession) getSysConfDir() string {
	return s.sysConfDir
}

func (s mirachSession) getUserConfDir() string {
	return s.userConfDir
}

func (s mirachSession) getConfigDirs() []string {
	return s.configDirs
}

func (s mirachSession) getConfig() string {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	for _, d := range s.configDirs {
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

func (s mirachSession) parseCommand() {
	if _, err := flags.Parse(&opts); err != nil {
		flagsErr, ok := err.(*flags.Error)
		if ok && flagsErr.Type == flags.ErrHelp {
			return
		}
		panic(flagsErr)
	}
}

func (s mirachSession) configureLogging() {
	s.verbosity = len(opts.Verbose)
	switch {
	case s.verbosity == 1:
		jww.SetStdoutThreshold(jww.LevelInfo)
	case s.verbosity > 1:
		jww.SetStdoutThreshold(jww.LevelTrace)
	}
}

func (s mirachSession) setConfigDirs() {
	if runtime.GOOS == "windows" {
		s.userConfDir = filepath.Join("%APPDATA%", "mirach")
		s.sysConfDir = filepath.Join("%PROGRAMDATA%", "mirach")
	} else {
		s.userConfDir = "$HOME/.config/mirach"
		s.sysConfDir = "/etc/mirach/"
	}
	s.configDirs = append(s.configDirs, ".", s.userConfDir, s.sysConfDir)
}

func (s mirachSession) initializeConfigAndLogging() {
	s.configureLogging()
	s.setConfigDirs()
	s.getConfig()
}

func (s mirachSession) getCustomer() *Customer {
	cust := new(Customer)
	err := cust.Init()
	if err != nil {
		msg := "customer initialization failed"
		customOut(msg, err)
		os.Exit(1)
	}
	s.customer = cust
	return s.customer
}

func (s mirachSession) getAsset(cust *Customer) *Asset {
	asset := new(Asset)
	err := asset.Init(cust)
	if err != nil {
		msg := "asset initialization failed"
		customOut(msg, err)
		os.Exit(1)
	}
	s.asset = asset
	return s.asset
}

func (s mirachSession) handlePlugins(client mqtt.Client, cron *cron.Cron) {
	plugins := make(map[string]Plugin)
	err := viper.UnmarshalKey("plugins", &plugins)
	if err != nil {
		jww.ERROR.Println(err)
	}
	cron.Start()
	for k, v := range plugins {
		jww.INFO.Printf("Adding to plugin: %s", k)
		cron.AddFunc(v.Schedule, RunPlugin(v, client))
	}
}

func (s mirachSession) handleCommands(asset *Asset) {
	err := asset.readCmds()
	if err != nil {
		msg := "stopped receiving commands; stopping mirach"
		customOut(msg, err)
		os.Exit(1)
	}
	msg := "mirach entered running state; plugins loaded"
	customOut(msg, nil)

}

func main() {
	Mirach := mirachSession{}
	Mirach.parseCommand()
	if opts.Version {
		showVersion()
		return
	}
	Mirach.initializeConfigAndLogging()
	cust := Mirach.getCustomer()
	asset := Mirach.getAsset(cust)
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	cron := cron.New()
	Mirach.handlePlugins(asset.client, cron)
	Mirach.handleCommands(asset)
	for _ = range signalChannel {
		// sig is a ^c, handle it
		jww.DEBUG.Println("SIGINT, stopping")
		cron.Stop()
		os.Exit(1)
	}
}
