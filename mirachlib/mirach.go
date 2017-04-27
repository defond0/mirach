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

	"gitlab.eng.cleardata.com/dash/mirach/cron"
	"gitlab.eng.cleardata.com/dash/mirach/plugin/envinfo"
	"gitlab.eng.cleardata.com/dash/mirach/util"

	jww "github.com/spf13/jwalterweatherman"
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

func getAsset() (*Asset, error) {
	asset := new(Asset)
	err := asset.Init()
	if err != nil {
		msg := "asset initialization failed"
		util.CustomOut(msg, err)
		return nil, err
	}
	return asset, nil
}

func handleCommands(asset *Asset) {
	err := asset.readCmds()
	if err != nil {
		msg := "stopped receiving commands; stopping mirach"
		util.CustomOut(msg, err)
		os.Exit(1)
	}
	msg := "mirach entered running state; plugins loaded"
	util.CustomOut(msg, nil)
}

// logResChan logs a result or error as return to the given channel.
// With this you can create a a chan to pass to a go routine then invoke this
// function. When the go routine to which the given channel was passed, the
// result will be logged.
func logResChan(successMsg, errMsg string, res chan interface{}) {
	switch r := <-res; r.(type) {
	case nil:
		jww.INFO.Println(successMsg)
	case string:
		jww.INFO.Println(successMsg + ": " + r.(string))
	case error:
		msg := fmt.Sprintf("go routine experienced error: %s", r.(error).Error())
		util.CustomOut(msg, r)
	default:
		err := fmt.Errorf("unexpected type in result chan")
		util.CustomOut(nil, err)
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
	handlePlugins(asset, cron)
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
