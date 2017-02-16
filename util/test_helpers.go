// +build integration unit

package util

import (
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/afero"
)

var osFs = afero.NewOsFs()
var testFs = afero.NewMemMapFs()

func writeTestCerts() {
	ca, err := afero.ReadFile(osFs, "../test_resources/ca.pem")
	if err != nil {
		panic(err)
	}
	testCA, err := testFs.Create("/etc/mirach/ca.pem")
	if err != nil {
		panic(err)
	}
	defer testCA.Close()
	_, err = testCA.WriteString(string(ca))
	if err != nil {
		panic(err)
	}
	// Write a keys to the $HOME directory to verify FindInDirs.
	privKey, err := afero.ReadFile(osFs, "../test_resources/customer/keys/private.pem.key")
	if err != nil {
		panic(err)
	}
	home, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	testPrivKey, err := testFs.Create(filepath.Join(home, ".config/mirach/customer/keys/private.pem.key"))
	if err != nil {
		panic(err)
	}
	defer testPrivKey.Close()
	_, err = testPrivKey.WriteString(string(privKey))
	if err != nil {
		panic(err)
	}
}
