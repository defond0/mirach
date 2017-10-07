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

// WriteTestCerts writes certs to the test filesystem for testing.
func WriteTestCerts() {
	ca, err := afero.ReadFile(OSFs, "../test_resources/ca.pem")
	if err != nil {
		panic(err)
	}
	privKey, err := afero.ReadFile(OSFs, "../test_resources/customer/keys/private.pem.key")
	if err != nil {
		panic(err)
	}
	cert, err := afero.ReadFile(OSFs, "../test_resources/customer/keys/ca.pem.crt")
	if err != nil {
		panic(err)
	}
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
	_, err = testCA.WriteString(string(ca))
	if err != nil {
		panic(err)
	}
	_, err = testPrivKey.WriteString(string(privKey))
	if err != nil {
		panic(err)
	}
	_, err = testCert.WriteString(string(cert))
	if err != nil {
		panic(err)
	}
}

func writeTestConfig(path string) {
	config, err := afero.ReadFile(OSFs, path)
	if err != nil {
		panic(err)
	}
	testConfig, err := TestFs.Create("/etc/mirach/config.yaml")
	if err != nil {
		panic(err)
	}
	defer testConfig.Close()
	_, err = testConfig.WriteString(string(config))
	if err != nil {
		panic(err)
	}
}

// WriteTestEvilConfig writes a known bad configuration for testing.
func WriteTestEvilConfig() {
	writeTestConfig("../test_resources/eve_config.yaml")
}

// WriteTestGoodConfig writes a known good configuration for testing.
func WriteTestGoodConfig() {
	writeTestConfig("../test_resources/config.yaml")
}
