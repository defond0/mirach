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
	logLevel    string
)

func configureLogging() {
	switch logLevel {
	case "info":
		jww.SetStdoutThreshold(jww.LevelInfo)
	case "trace":
		jww.SetStdoutThreshold(jww.LevelTrace)
	}
}

// CustomOut either outputs feedback or a log message at error level.
func CustomOut(fbMsg, err interface{}) {
	switch logLevel {
	case "info", "trace":
		if err != nil {
			jww.ERROR.Println(fmt.Sprint(err))
		} else {
			jww.INFO.Println(fmt.Sprint(fbMsg))
		}
	default:
		jww.FEEDBACK.Println(fmt.Sprint(fbMsg))
	}
}

func getAsset(cust *Customer) (*Asset, error) {
	asset := new(Asset)
	err := asset.Init(cust)
	if err != nil {
		msg := "asset initialization failed"
		CustomOut(msg, err)
		return nil, err
	}
	return asset, nil
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

func getCustomer() (*Customer, error) {
	cust := new(Customer)
	err := cust.Init()
	if err != nil {
		msg := "customer initialization failed"
		CustomOut(msg, err)
		return nil, err
	}
	return cust, nil
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

// PrepResources set up requirements and returns nodes.
func PrepResources() (*Customer, *Asset, error) {
	configureLogging()
	getConfig()
	cust, err := getCustomer()
	if err != nil {
		return nil, nil, err
	}
	asset, err := getAsset(cust)
	if err != nil {
		return nil, nil, err
	}
	return cust, asset, nil
}

// RunLoop begins the long running portions of the application.
// This function will run indefinitely. It creates and manages the cron scheduler.
// It also calls for the initialization of clients and signal channels.
func RunLoop(cust *Customer, asset *Asset) {
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

// SetLogLevel sets the log level variable.
func SetLogLevel(level string) error {
	levels := []string{"error", "info", "trace"}
	for _, l := range levels {
		if level == l {
			logLevel = l
			return nil
		}
	}
	return fmt.Errorf("choose level from %s", levels)
}

// Start is the main entry for the mirachlib.
func Start() error {
	cust, asset, err := PrepResources()
	if err != nil {
		return err
	}
	RunLoop(cust, asset)
	return nil
}
