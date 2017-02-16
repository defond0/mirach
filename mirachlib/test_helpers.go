// +build integration unit

package mirachlib

import "github.com/spf13/afero"

var osFs = afero.NewOsFs()
var testFs = afero.NewMemMapFs()

func resetTestFS() {
	testFs = afero.NewMemMapFs()
}

func writeTestCerts() {
	ca, err := afero.ReadFile(osFs, "../test_resources/ca.pem")
	if err != nil {
		panic(err)
	}
	privKey, err := afero.ReadFile(osFs, "../test_resources/customer/keys/private.pem.key")
	if err != nil {
		panic(err)
	}
	cert, err := afero.ReadFile(osFs, "../test_resources/customer/keys/ca.pem.crt")
	if err != nil {
		panic(err)
	}
	testCA, err := testFs.Create("/etc/mirach/ca.pem")
	if err != nil {
		panic(err)
	}
	defer testCA.Close()
	testPrivKey, err := testFs.Create("/etc/mirach/customer/keys/private.pem.key")
	if err != nil {
		panic(err)
	}
	defer testPrivKey.Close()
	testCert, err := testFs.Create("/etc/mirach/customer/keys/ca.pem.crt")
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
	config, err := afero.ReadFile(osFs, path)
	if err != nil {
		panic(err)
	}
	testConfig, err := testFs.Create("/etc/mirach/config.yaml")
	if err != nil {
		panic(err)
	}
	defer testConfig.Close()
	_, err = testConfig.WriteString(string(config))
	if err != nil {
		panic(err)
	}
}

func writeTestEvilConfig() {
	writeTestConfig("../test_resources/eve_config.yaml")
}

func writeTestGoodConfig() {
	writeTestConfig("../test_resources/config.yaml")
}
