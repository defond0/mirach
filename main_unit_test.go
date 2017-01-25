// +build unit

package main

import (
	"testing"

	jww "github.com/spf13/jwalterweatherman"
	"github.com/stretchr/testify/assert"
)

func TestMainConfigureLogging(t *testing.T) {
	Mirach = &mirachSession{}
	assert := assert.New(t)
	opts.Verbose = []bool{}
	Mirach.configureLogging()
	assert.Equal(Mirach.getVerbosity(), 0, "Expected vebosity of 0")
	assert.Equal(jww.StdoutThreshold(), jww.LevelError)
	opts.Verbose = []bool{true}
	Mirach.configureLogging()
	assert.Equal(Mirach.getVerbosity(), 1, "Expected vebosity of 1")
	assert.Equal(jww.StdoutThreshold(), jww.LevelInfo)
	opts.Verbose = []bool{true, true, true}
	Mirach.configureLogging()
	assert.Equal(Mirach.getVerbosity(), 3, "Expected vebosity of 3")
	assert.Equal(jww.StdoutThreshold(), jww.LevelTrace)
}

func TestSetConfigDirs(t *testing.T) {
	// Just checking that configDirs is correct amalgam
	Mirach = &mirachSession{}
	assert := assert.New(t)
	Mirach.setConfigDirs()
	assert.NotNil(Mirach.getSysConfDir())
	assert.NotNil(Mirach.getUserConfDir())
	assert.Equal(Mirach.getConfigDirs(), []string{".", Mirach.getUserConfDir(), Mirach.getSysConfDir()})
}

// func TestRunPlugin(t *testing.T) {
// 	viper.SetConfigName("config")
// 	viper.Set("customer.id", "unit-test")
// 	viper.Set("asset.id", "unit-test")
// 	shell("cp", "./test_resources/unit-test.sh", ".")
// 	defer shell("rm", "./unit-test.sh")
// 	assert := assert.New(t)
// 	test_plug := Plugin{}
// 	test_plug.Label = "unit-test-plugin"
// 	test_plug.Cmd = "./unit-test.sh"
// 	test_plug.Schedule = "doesn't matter"
// 	// c := new(mocks.Client)
// 	// token := new(mocks.Token)
// 	// token.On("Wait").Return()
// 	// var qos uint8
// 	// qos = 1
// 	// c.On("Publish", "mirach/data/unit-test/unit-test", qos, false, "{\"type\":\"unit\",\"data\":\"test\"}").Return(token)
// 	// out := RunPlugin(test_plug, c)
// 	// assert.NotNil(out)
// 	// out()
// 	// token.AssertExpectations(t)
// 	// c.AssertExpectations(t)
// }
