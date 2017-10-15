// Package lib provides the main application components of mirach.
//
// mirachlib is inextricably linked to the util package. If during testing
// you use any function from the util package that operates with files
// you will need to use the util.SetFs function to use an in-memory filesystem.
package lib

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"reflect"

	"github.com/cleardataeng/mirach/lib/cron"
	"github.com/cleardataeng/mirach/lib/input"
	"github.com/cleardataeng/mirach/lib/util"

	jww "github.com/spf13/jwalterweatherman"
	"github.com/theherk/viper"
)

func backends() []io.ReadWriter {
	return []io.ReadWriter{}
}

// RunLoop begins the long running portions of the application.
// This function will run indefinitely. It creates and manages the cron scheduler.
// It also calls for the initialization of clients and signal channels.
func RunLoop() {
	defer func() {
		if r := recover(); r != nil {
			if reflect.TypeOf(r).String() == "plugin.Exception" {
				util.CustomOut("error in plugin (restarting)", r)
				RunLoop()
			}
			fmt.Println(r)
			os.Exit(1)
		}
	}()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	cron := cron.New()
	cron.Start()
	plugin.NewCtrl(cron, backends()...)
	for _ = range sigChan {
		// sig is a ^c, handle it
		jww.DEBUG.Println("SIGINT, stopping")
		Cron.Stop()
		os.Exit(1)
	}
}

// Start is the main entry for the mirachlib.
func Start() error {
	if _, err := util.GetConfig(); err != nil {
		cfgType := input.ReadCfgType()
		if err := util.BlankConfig(cfgType); err != nil {
			return err
		}
	}
	switch viper.GetString("log_level") {
	case "info":
		jww.SetStdoutThreshold(jww.LevelInfo)
	case "trace":
		jww.SetStdoutThreshold(jww.LevelTrace)
	}
	RunLoop()
	return nil
}
