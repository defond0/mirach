// +build unit

package mirachlib

import (
	"testing"

	"github.com/cleardataeng/mirach/util"

	jww "github.com/spf13/jwalterweatherman"
	"github.com/stretchr/testify/assert"
)

func TestConfigureLogging(t *testing.T) {
	assert := assert.New(t)
	util.SetFs(util.TestFs)
	util.WriteTestCerts()
	util.WriteTestGoodConfig()
	configureLogging()
	assert.Equal("error", logLevel, "log level default error")
	assert.Equal(jww.LevelError, jww.StdoutThreshold())
}
