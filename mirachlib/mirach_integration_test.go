// +build integration

package mirachlib

import (
	"flag"
	"os"
	"testing"

	"gitlab.eng.cleardata.com/dash/mirach/util"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func TestIntegrationMainRegistration(t *testing.T) {
	assert := assert.New(t)
	writeTestCerts()
	writeTestGoodConfig()
	util.SetFs(testFs)
	PrepResources()
	// Verify we received correct customer_id and wrote it to config.
	v := viper.New()
	v.AddConfigPath("/etc/mirach/")
	v.SetConfigName("config")
	v.SetFs(testFs)
	err := v.ReadInConfig()
	assert.Nil(err)
	assert.Equal("00000666", v.GetString("customer.id"), "value in read config")
	// Verify the private key file was written and can be read.
	priv, err := util.ReadFile("/etc/mirach/asset/keys/private.pem.key")
	assert.Nil(err)
	assert.NotEmpty(priv)
	// Verify the cert was written and can be read.
	cert, err := util.ReadFile("/etc/mirach/asset/keys/ca.pem.crt")
	assert.Nil(err)
	assert.NotEmpty(cert)
}

//Attempt to register with customer number 00006913(not it's own)
func TestIntegrationMainEvilListener(t *testing.T) {
	assert := assert.New(t)
	resetTestFS()
	util.SetFs(testFs)
	writeTestCerts()
	writeTestEvilConfig()
	PrepResources()
	// Verify we used our evil config.
	v := viper.New()
	v.AddConfigPath("/etc/mirach/")
	v.SetConfigName("config")
	v.SetFs(testFs)
	err := v.ReadInConfig()
	assert.Equal("00006913", v.GetString("customer.id"), "value in read config")
	priv, err := util.ReadFile("/etc/mirach/asset/keys/private.pem.key")
	assert.NotNil(err)
	assert.Empty(priv)
	cert, err := util.ReadFile("/etc/mirach/asset/keys/ca.pem.crt")
	assert.NotNil(err)
	assert.Empty(cert)
}
