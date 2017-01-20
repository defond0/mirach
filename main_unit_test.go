// +build unit

package main

import (
	"testing"

	jww "github.com/spf13/jwalterweatherman"
	"github.com/stretchr/testify/assert"
)

func TestMainConfigureLogging(t *testing.T) {
	assert := assert.New(t)
	opts.Verbose = []bool{}
	configureLogging()
	assert.Equal(AppConfig.verbosity, 0, "Expected vebosity of 0")
	assert.Equal(jww.StdoutThreshold(), jww.LevelError)
	opts.Verbose = []bool{true}
	configureLogging()
	assert.Equal(AppConfig.verbosity, 1, "Expected vebosity of 1")
	assert.Equal(jww.StdoutThreshold(), jww.LevelInfo)
	opts.Verbose = []bool{true, true, true}
	configureLogging()
	assert.Equal(AppConfig.verbosity, 3, "Expected vebosity of 3")
	assert.Equal(jww.StdoutThreshold(), jww.LevelTrace)
}

func TestSetConfigDirs(t *testing.T) {
	// Just checking that configDirs is correct amalgam
	assert := assert.New(t)
	setConfigDirs()
	assert.NotNil(AppConfig.sysConfDir)
	assert.NotNil(AppConfig.userConfDir)
	assert.Equal(AppConfig.configDirs, []string{".", AppConfig.userConfDir, AppConfig.sysConfDir})
}

// func TestInitializeConfigAndLogging(t *testing.T) {
// 	assert := assert.New(t)
// 	configCalls := -1
// 	getConfig := func() {
// 		configCalls += 1
// 	}
// 	getConfig()
// 	loggingCalls := -1
// 	configureLogging := func() {
// 		loggingCalls += 1
// 	}
// 	configureLogging()
// 	confDirCalls := -1
// 	setConfigDirs := func() {
// 		confDirCalls += 1
// 	}
// 	setConfigDirs()

// 	initializeConfigAndLogging()
// 	assert.Equal(loggingCalls, 1)
// 	assert.Equal(confDirCalls, 1)
// 	assert.Equal(configCalls, 1)
// }
