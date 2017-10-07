// +build integration unit

package util

import "github.com/spf13/afero"

// OSFs is the operating system's filesystem.
var OSFs = afero.NewOsFs()

// TestFs is the in-memory filesystem used for testing.
var TestFs = afero.NewMemMapFs()

// ResetTestFs creates a new TestFs blank slate.
func ResetTestFs() {
	TestFs = afero.NewMemMapFs()
}

var (
	ca      = `your cert here` + "\n"
	cert    = `test ca.pem.crt contents` + "\n"
	privKey = `test private.pem.key` + "\n"
)

var eveCfg = `asset:
  id: "000006913-evil-and-bad-for-you"
customer:
  id: "00006913"
`

var goodCfg = `asset:
  id: "00000666-mirach"
`

// WriteTestCerts writes certs to the test filesystem for testing.
func WriteTestCerts() {
	testCA, err := TestFs.Create("/etc/mirach/ca.pem")
	if err != nil {
		panic(err)
	}
	defer testCA.Close()
	testPrivKey, err := TestFs.Create("/etc/mirach/customer/keys/private.pem.key")
	if err != nil {
		panic(err)
	}
	defer testPrivKey.Close()
	testCert, err := TestFs.Create("/etc/mirach/customer/keys/ca.pem.crt")
	if err != nil {
		panic(err)
	}
	defer testCert.Close()
	_, err = testCA.WriteString(ca)
	if err != nil {
		panic(err)
	}
	_, err = testPrivKey.WriteString(privKey)
	if err != nil {
		panic(err)
	}
	_, err = testCert.WriteString(cert)
	if err != nil {
		panic(err)
	}
}

func writeTestConfig(config string) {
	testConfig, err := TestFs.Create("/etc/mirach/config.yaml")
	if err != nil {
		panic(err)
	}
	defer testConfig.Close()
	_, err = testConfig.WriteString(config)
	if err != nil {
		panic(err)
	}
}

// WriteTestEvilConfig writes a known bad configuration for testing.
func WriteTestEvilConfig() {
	writeTestConfig(eveCfg)
}

// WriteTestGoodConfig writes a known good configuration for testing.
func WriteTestGoodConfig() {
	writeTestConfig(goodCfg)
}
