// +build unit

package mirachlib

import (
	"testing"

	"cleardata.com/dash/mirach/util"

	jww "github.com/spf13/jwalterweatherman"
	"github.com/stretchr/testify/assert"
)

func TestConfigureLogging(t *testing.T) {
	assert := assert.New(t)
	util.SetFs(testFs)
	writeTestCerts()
	writeTestGoodConfig()
	configureLogging()
	assert.Equal("error", logLevel, "log level default error")
	assert.Equal(jww.LevelError, jww.StdoutThreshold())
}

func TestGetConfig(t *testing.T) {
	assert := assert.New(t)
	util.SetFs(testFs)
	writeTestCerts()
	writeTestGoodConfig()
	getConfig()
	assert.NotNil(confDirs)
	assert.NotNil(userConfDir)
	assert.NotNil(sysConfDir)
	assert.Len(confDirs, 3, "current, user, and system")
}
