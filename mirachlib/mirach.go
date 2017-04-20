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
	"time"

	"gitlab.eng.cleardata.com/dash/mirach/cron"
	"gitlab.eng.cleardata.com/dash/mirach/plugin/compinfo"
	"gitlab.eng.cleardata.com/dash/mirach/plugin/ebsinfo"
	"gitlab.eng.cleardata.com/dash/mirach/plugin/envinfo"
	"gitlab.eng.cleardata.com/dash/mirach/plugin/pkginfo"
	"gitlab.eng.cleardata.com/dash/mirach/util"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/theherk/viper"
)

var (
	confDirs    []string
	sysConfDir  string
	userConfDir string
	logLevel    = "error" // default log level
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

func getAsset() (*Asset, error) {
	asset := new(Asset)
	err := asset.Init()
	if err != nil {
		msg := "asset initialization failed"
		CustomOut(msg, err)
		return nil, err
	}
	return asset, nil
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

func handlePlugins(client mqtt.Client, cron *cron.MirachCron) {
	externalPlugins := make(map[string]ExternalPlugin)
	err := viper.UnmarshalKey("plugins", &externalPlugins)
	if err != nil {
		jww.ERROR.Println(err)
	}
	internalPlugins := []InternalPlugin{
		{
			Label:    "compinfo-docker",
			Schedule: "@hourly",
			StrFunc:  compinfo.GetDockerString,
			Type:     "compinfo",
		},
		{
			Label:    "compinfo-load",
			Schedule: "@every 5m",
			StrFunc:  compinfo.GetLoadString,
			Type:     "compinfo",
		},
		{
			Label:    "compinfo-sys",
			Schedule: "@daily",
			StrFunc:  compinfo.GetSysString,
			Type:     "compinfo",
		},
		{
			Label:    "pkginfo",
			Schedule: "@daily",
			StrFunc:  pkginfo.String,
			Type:     "pkginfo",
		},
	}
	if envinfo.Env.CloudProvider == "aws" {
		AWSPlugins := []InternalPlugin{
			{
				Label:    "ebsinfo",
				Schedule: "@daily",
				StrFunc:  ebsinfo.String,
				Type:     "ebsinfo",
			},
		}
		internalPlugins = append(internalPlugins, AWSPlugins...)
	}
	cron.Start()
	for k, v := range externalPlugins {
		// Loop over internal plugins to check name collisions.
		ok := true
		for _, p := range internalPlugins {
			if v.Label == p.Label || v.Label == p.Type {
				err = fmt.Errorf("refusing to load plugin %v: internal name taken", k)
				CustomOut(nil, err)
				ok = false
				break
			}
		}
		if !ok {
			continue
		}
		delay, err := time.ParseDuration(v.LoadDelay)
		if err != nil {
			if err.Error() == "time: invalid duration" {
				jww.INFO.Printf("adding plugin to cron: %s", k)
				err := cron.AddFunc(v.Schedule, v.Run(client))
				if err != nil {
					msg := fmt.Sprintf("failed to load plugin %v", k)
					CustomOut(msg, err)
				}
			}
			msg := fmt.Sprintf("failed to parse delay during load plugin %v", k)
			CustomOut(msg, err)
		}
		jww.INFO.Printf("adding plugin: %s to cron with start delay: %s", k, delay)
		res := make(chan interface{})
		cron.AddFuncDelayed(v.Schedule, v.Run(client), delay, res)
		successMsg := fmt.Sprintf("added plugin: %s to cron after: %s", k, delay)
		errorMsg := fmt.Sprintf("failed to load plugin: %s to cron after: %s", k, delay)
		go logResChan(successMsg, errorMsg, res)
	}
	for _, v := range internalPlugins {
		jww.INFO.Printf("adding plugin: %s to cron with start delay: %s", v.Label, v.LoadDelay)
		res := make(chan interface{})
		cron.AddFuncDelayed(v.Schedule, v.Run(client), v.LoadDelay, res)
		successMsg := fmt.Sprintf("added plugin: %s to cron after: %s", v.Label, v.LoadDelay)
		errorMsg := fmt.Sprintf("failed to load plugin: %s to cron after: %s", v.Label, v.LoadDelay)
		go logResChan(successMsg, errorMsg, res)
	}
}

// logResChan logs a result or error as return to the given channel.
// With this you can create a a chan to pass to a go routine then invoke this
// function. When the go routine to which the given channel was passed, the
// result will be logged.
func logResChan(successMsg, errMsg string, res chan interface{}) {
	switch r := <-res; r.(type) {
	case nil:
		CustomOut(successMsg, nil)
	case string:
		CustomOut(successMsg+": "+r.(string), nil)
	case error:
		msg := fmt.Sprintf("go routine experienced error: %s", r.(error).Error())
		CustomOut(msg, r)
	default:
		err := fmt.Errorf("unexpected type in result chan")
		CustomOut(nil, err)
	}
}

// PrepResources set up requirements and returns nodes.
func PrepResources() (*Asset, error) {
	var err error
	configureLogging()
	confDirs, err = util.GetConfDirs()
	if err != nil {
		return nil, err
	}
	userConfDir, sysConfDir = confDirs[1], confDirs[2]
	_, err = util.GetConfig(confDirs)
	if err != nil {
		cfgType := readCfgType()
		err := util.BlankConfig(cfgType, sysConfDir)
		if err != nil {
			return nil, err
		}
	}
	asset, err := getAsset()
	if err != nil {
		return nil, err
	}
	return asset, nil
}

// RunLoop begins the long running portions of the application.
// This function will run indefinitely. It creates and manages the cron scheduler.
// It also calls for the initialization of clients and signal channels.
func RunLoop(asset *Asset) {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	cron := cron.New()
	if envinfo.Env == nil {
		envinfo.Env = new(envinfo.EnvInfoGroup)
		envinfo.Env.GetInfo()
	}
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
	asset, err := PrepResources()
	if err != nil {
		return err
	}
	RunLoop(asset)
	return nil
}
