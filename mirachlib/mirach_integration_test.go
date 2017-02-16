// +build integration

package mirachlib

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/spf13/afero"
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "../mirach")
	stdoutpipe, _ := cmd.StdoutPipe()
	stdoutscanner := bufio.NewScanner(stdoutpipe)
	if err := cmd.Start(); err != nil {
		fmt.Println(err)
	}
	for stdoutscanner.Scan() {
		scan := stdoutscanner.Text()
		fmt.Println(scan)
		if scan == "mirach entered running state; plugins loaded" {
			cancel()
			// Verify we received correct customer_id and wrote it to config.
			v := viper.New()
			v.AddConfigPath("/etc/mirach/")
			v.SetConfigName("config")
			v.SetFs(testFs)
			err := v.ReadInConfig()
			assert.Nil(err)
			assert.Equal("00000666", v.GetString("customer.id"), "value in read config")
			// Verify the private key file was written and can be read.
			priv, err := afero.ReadFile(testFs, "/etc/mirach/asset/keys/private.pem.key")
			assert.Nil(err)
			assert.NotEmpty(priv)
			// Verify the cert was written and can be read.
			cert, err := afero.ReadFile(testFs, "/etc/mirach/asset/keys/ca.pem.crt")
			assert.Nil(err)
			assert.NotEmpty(cert)
		}
	}
	select {
	case <-ctx.Done():
		assert.Equal("context canceled", fmt.Sprint(ctx.Err()))
	}
}

func TestIntegrationMainRegistrationOld(t *testing.T) {
	assert := assert.New(t)
	writeTestCerts()
	writeTestGoodConfig()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "../mirach")
	stdoutpipe, _ := cmd.StdoutPipe()
	stdoutscanner := bufio.NewScanner(stdoutpipe)
	if err := cmd.Start(); err != nil {
		fmt.Println(err)
	}
	for stdoutscanner.Scan() {
		scan := stdoutscanner.Text()
		fmt.Println(scan)
		if scan == "mirach entered running state; plugins loaded" {
			cancel()
			// Verify we received correct customer_id and wrote it to config.
			v := viper.New()
			v.AddConfigPath("/etc/mirach/")
			v.SetConfigName("config")
			v.SetFs(testFs)
			err := v.ReadInConfig()
			assert.Nil(err)
			assert.Equal("00000666", v.GetString("customer.id"), "value in read config")
			// Verify the private key file was written and can be read.
			priv, err := afero.ReadFile(testFs, "/etc/mirach/asset/keys/private.pem.key")
			assert.Nil(err)
			assert.NotEmpty(priv)
			// Verify the cert was written and can be read.
			cert, err := afero.ReadFile(testFs, "/etc/mirach/asset/keys/ca.pem.crt")
			assert.Nil(err)
			assert.NotEmpty(cert)
		}
	}
	select {
	case <-ctx.Done():
		assert.Equal("context canceled", fmt.Sprint(ctx.Err()))
	}
}

//Attempt to register with customer number 00006913(not it's own)
func TestIntegrationMainEvilListener(t *testing.T) {
	assert := assert.New(t)
	writeTestCerts()
	writeTestEvilConfig()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "../mirach")
	stdoutpipe, _ := cmd.StdoutPipe()
	stdoutscanner := bufio.NewScanner(stdoutpipe)
	if err := cmd.Start(); err != nil {
		fmt.Println(err)
	}
	for stdoutscanner.Scan() {
		scan := stdoutscanner.Text()
		if scan == "asset initialization failed" {
			cancel()
			// Verify we used our evil config.
			assert.Equal("00006913", viper.GetString("customer.id"), "value in read config")
			priv, err := afero.ReadFile(testFs, "/etc/mirach/asset/keys/private.pem.key")
			assert.NotNil(err)
			assert.Empty(priv)
			cert, err := ioutil.ReadFile("/etc/mirach/asset/keys/ca.pem.crt")
			assert.NotNil(err)
			assert.Empty(cert)
		}
	}
	select {
	case <-ctx.Done():
		assert.Equal("context canceled", fmt.Sprint(ctx.Err()))
	}

}
