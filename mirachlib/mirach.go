// Package mirachlib provides the main application components of mirach.
//
// mirachlib is inextricably linked to the util package. If during testing
// you use any function from the util package that operates with files
// you will need to use the util.SetFs function to use an in-memory filesystem.
package mirachlib

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"

	"cleardata.com/dash/mirach/util"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/robfig/cron"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

var (
	confDirs    []string
	sysConfDir  string
	userConfDir string
	verbosity   int
)

func configureLogging() {
	switch {
	case verbosity == 1:
		jww.SetStdoutThreshold(jww.LevelInfo)
	case verbosity > 1:
		jww.SetStdoutThreshold(jww.LevelTrace)
	}
}

// CustomOut either outputs feedback or a log message at error level.
func CustomOut(fbMsg, err interface{}) {
	switch {
	case verbosity > 0:
		if err != nil {
			jww.ERROR.Println(fmt.Sprint(err))
		} else {
			jww.INFO.Println(fmt.Sprint(fbMsg))
		}
	default:
		jww.FEEDBACK.Println(fmt.Sprint(fbMsg))
	}
}

func getAsset(cust *Customer) *Asset {
	asset := new(Asset)
	err := asset.Init(cust)
	if err != nil {
		msg := "asset initialization failed"
		CustomOut(msg, err)
		os.Exit(1)
	}
	return asset
}

func getConfig() string {
	if len(confDirs) == 0 {
		if runtime.GOOS == "windows" {
			// TODO: Will probably need to populate these with github.com/luisiturrios/gowin.
			// Currently gowin is failing to install. Tracking with luisiturrios/gowin#5.
			userConfDir = filepath.Join("%APPDATA%", "mirach")
			sysConfDir = filepath.Join("%PROGRAMDATA%", "mirach")
		} else {
			home, err := homedir.Dir()
			if err != nil {
				panic(err)
			}
			userConfDir = filepath.Join(home, ".config/mirach")
			sysConfDir = "/etc/mirach/"
		}
		confDirs = append(confDirs, ".", userConfDir, sysConfDir)
	}
	viper.SetConfigName("config")
	for _, d := range confDirs {
		viper.AddConfigPath(d)
	}
	viper.SetFs(util.Fs)
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	viper.SetEnvPrefix("mirach")
	viper.AutomaticEnv()
	return viper.ConfigFileUsed()
}

func getCustomer() *Customer {
	cust := new(Customer)
	err := cust.Init()
	if err != nil {
		msg := "customer initialization failed"
		CustomOut(msg, err)
		os.Exit(1)
	}
	return cust
}

func handleCommands(asset *Asset) {
	err := asset.readCmds()
	if err != nil {
		msg := "stopped receiving commands; stopping mirach"
		CustomOut(msg, err)
		os.Exit(1)
	}
	msg := "mirach entered running state; plugins loaded"
	CustomOut(msg, nil)
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
		err := cron.AddFunc(v.Schedule, RunPlugin(v, client))
		if err != nil {
			msg := "failed to launch plugins"
			CustomOut(msg, err)
			os.Exit(1)
		}
	}
}

// Start begins the application.
// This function will run indefinitely. It creates and manages the cron scheduler.
// It also calls for the initialization of clients and signal channels.
func Start() {
	configureLogging()
	getConfig()
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
