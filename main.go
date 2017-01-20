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

var AppConfig struct {
	sysConfDir  string
	userConfDir string
	configDirs  []string
	verbosity   int
}
var opts struct {
	Verbose []bool `short:"v" long:"verbose" description:"Show verbose debug information"`
	Version bool   `long:"version" description:"Show version"`
}

func getConfig() string {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	for _, d := range AppConfig.configDirs {
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

func parseCommand() {
	if _, err := flags.Parse(&opts); err != nil {
		flagsErr, ok := err.(*flags.Error)
		if ok && flagsErr.Type == flags.ErrHelp {
			return
		}
		panic(flagsErr)
	}
}

func configureLogging() {
	AppConfig.verbosity = len(opts.Verbose)
	switch {
	case AppConfig.verbosity == 1:
		jww.SetStdoutThreshold(jww.LevelInfo)
	case AppConfig.verbosity > 1:
		jww.SetStdoutThreshold(jww.LevelTrace)
	}
}

func setConfigDirs() {
	if runtime.GOOS == "windows" {
		AppConfig.userConfDir = filepath.Join("%APPDATA%", "mirach")
		AppConfig.sysConfDir = filepath.Join("%PROGRAMDATA%", "mirach")
	} else {
		AppConfig.userConfDir = "$HOME/.config/mirach"
		AppConfig.sysConfDir = "/etc/mirach/"
	}
	AppConfig.configDirs = append(AppConfig.configDirs, ".", AppConfig.userConfDir, AppConfig.sysConfDir)
}

func initializeConfigAndLogging() {
	configureLogging()
	setConfigDirs()
	getConfig()
}

func getCustomer() *Customer {
	cust := new(Customer)
	err := cust.Init()
	if err != nil {
		msg := "customer initialization failed"
		customOut(msg, err)
		os.Exit(1)
	}
	return cust
}

func getAsset(cust *Customer) *Asset {
	asset := new(Asset)
	err := asset.Init(cust)
	if err != nil {
		msg := "asset initialization failed"
		customOut(msg, err)
		os.Exit(1)
	}
	return asset
}

func handlePlugins(client mqtt.Client, cron *cron.Cron) {
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

func handleCommands(asset *Asset) {
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
	parseCommand()
	if opts.Version {
		showVersion()
		return
	}
	initializeConfigAndLogging()
	cust := getCustomer()
	asset := getAsset(cust)
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	cron := cron.New()
	handlePlugins(asset.client, cron)
	handleCommands(asset)
	for _ = range signalChannel {
		// sig is a ^c, handle it
		jww.DEBUG.Println("SIGINT, stopping")
		cron.Stop()
		os.Exit(1)
	}
}
