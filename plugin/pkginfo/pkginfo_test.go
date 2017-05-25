// +build unit

package pkginfo

import (
	"testing"

	"github.com/cleardataeng/mirach/plugin/pkginfo/parsers"

	"encoding/json"
)

type MockInfoGroup struct {
	mockGetInfo func()
	mockString  func() string
}

func (m *MockInfoGroup) GetInfo() {
	m.mockGetInfo()
}

func (m *MockInfoGroup) String() string {
	return m.mockString()
}

func TestKbStatusGetString(t *testing.T) {
	testInfo := make(map[string]map[string]parsers.KBArticle)
	testInfo["installed"] = map[string]parsers.KBArticle{
		"KB666": {
			Security: false,
		},
	}
	testInfo["available"] = map[string]parsers.KBArticle{
		"KB667": {
			Security: true,
		},
	}
	testInfo["available_security"] = map[string]parsers.KBArticle{
		"KB667": {
			Security: true,
		},
	}
	ogKb := new(KBStatus)
	mockKb := &MockInfoGroup{
		mockGetInfo: func() {
			ogKb.Articles = testInfo
		},
		mockString: ogKb.String,
	}
	mockKb.GetInfo()
	newKb := new(KBStatus)
	if err := json.Unmarshal([]byte(mockKb.String()), &newKb); err != nil {
		t.Error("can't unmarshall kbs")
	}
	if !(newKb.Articles["available"]["KB667"] == ogKb.Articles["available"]["KB667"]) {
		t.Error("available kbs don't match")
	}
	if !(newKb.Articles["available_security"]["KB667"] == ogKb.Articles["available_security"]["KB667"]) {
		t.Error("available_security kbs don't match")
	}
	if !(newKb.Articles["installed"]["KB666"] == ogKb.Articles["installed"]["KB666"]) {
		t.Error("installed kbs don't match")
	}

}

func TestPkgStatusGetString(t *testing.T) {
	testInfo := make(map[string]map[string]parsers.LinuxPackage)
	testInfo["installed"] = map[string]parsers.LinuxPackage{
		"ssh": {
			Version:  "0.1.0",
			Security: false,
		},
	}
	testInfo["available"] = map[string]parsers.LinuxPackage{
		"ssh": {
			Version:  "0.1.2",
			Security: true,
		},
	}
	testInfo["available_security"] = map[string]parsers.LinuxPackage{
		"ssh": {
			Version:  "0.1.2",
			Security: true,
		},
	}
	ogPkgs := new(PkgStatus)
	mockPkgs := &MockInfoGroup{
		mockGetInfo: func() {
			ogPkgs.Packages = testInfo
		},
		mockString: ogPkgs.String,
	}
	mockPkgs.GetInfo()
	newPkgs := new(PkgStatus)
	if err := json.Unmarshal([]byte(mockPkgs.String()), &newPkgs); err != nil {
		t.Error("can't unmarshall pkgs")
	}
	if !(newPkgs.Packages["available"]["ssh"] == ogPkgs.Packages["available"]["ssh"]) {
		t.Error("available pkgs don't match")
	}
	if !(newPkgs.Packages["available_security"]["ssh"] == ogPkgs.Packages["available_security"]["ssh"]) {
		t.Error("available_security pkgs don't match")
	}
	if !(newPkgs.Packages["installed"]["ssh"] == ogPkgs.Packages["installed"]["ssh"]) {
		t.Error("installed pkgs don't match")
	}

}
