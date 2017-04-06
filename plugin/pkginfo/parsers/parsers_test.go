// +build unit

package parsers

import (
	"fmt"
	"testing"
)

func TestParsePackagesFromBytes(t *testing.T) {
	txt := `
openssl 12.2.amd64
ssh 14.1.amd64
`
	out := []byte(txt)
	pkgs, err := parsePacakgesFromBytes(out, false)
	if err != nil {
		t.Error(fmt.Sprintf("error to parsing %s", txt))
	}
	if len(pkgs) != 2 {
		t.Error("wrong number of packages parsed")
	}
	// strings are ordered, and lists are ordered this is not
	// a design feature, just happens to be they are in the list
	// in the same order as they are passed in
	if pkgs["openssl"].name != "openssl" || pkgs["openssl"].Version != "12.2.amd64" || pkgs["ssh"].name != "ssh" || pkgs["ssh"].Version != "14.1.amd64" || pkgs["openssl"].Security || pkgs["ssh"].Security {
		t.Error("we parsed the wrong info")
	}
}
