// +build unit

package pkginfo

import (
	"fmt"
	"testing"

	"gitlab.eng.cleardata.com/dash/mirach/plugin/pkginfo/parsers"

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
	testInfo := make(map[string][]parsers.KBArticle)
	testInfo["installed"] = []parsers.KBArticle{
		{
			Name:     "KB666",
			Security: false,
		},
	}
	testInfo["available"] = []parsers.KBArticle{
		{
			Name:     "KB667",
			Security: true,
		},
	}
	testInfo["available_security"] = []parsers.KBArticle{
		{
			Name:     "KB667",
			Security: true,
		},
	}
	ogKb := new(KBStatus)
	mockKb := &MockInfoGroup{
		mockGetInfo: func() {
			fmt.Println("Get mock")
			ogKb.Articles = testInfo
		},
		mockString: ogKb.String,
	}
	mockKb.GetInfo()
	newKb := new(KBStatus)
	if err := json.Unmarshal([]byte(mockKb.String()), &newKb); err != nil {
		t.Error("can't unmarshall kbs")
	}
	if !(newKb.Articles["available"][0] == ogKb.Articles["available"][0]) {
		t.Error("available kbs don't match")
	}
	if !(newKb.Articles["available_security"][0] == ogKb.Articles["available_security"][0]) {
		t.Error("available_security kbs don't match")
	}
	if !(newKb.Articles["installed"][0] == ogKb.Articles["installed"][0]) {
		t.Error("installed kbs don't match")
	}

}

func TestPkgStatusGetString(t *testing.T) {
	testInfo := make(map[string][]parsers.LinuxPackage)
	testInfo["installed"] = []parsers.LinuxPackage{
		{
			Name:     "ssh",
			Version:  "0.1.0",
			Security: false,
		},
	}
	testInfo["available"] = []parsers.LinuxPackage{
		{
			Name:     "ssh",
			Version:  "0.1.2",
			Security: true,
		},
	}
	testInfo["available_security"] = []parsers.LinuxPackage{
		{
			Name:     "ssh",
			Version:  "0.1.2",
			Security: true,
		},
	}
	ogKb := new(PkgStatus)
	mockKb := &MockInfoGroup{
		mockGetInfo: func() {
			fmt.Println("Get mock")
			ogKb.Packages = testInfo
		},
		mockString: ogKb.String,
	}
	mockKb.GetInfo()
	newKb := new(PkgStatus)
	if err := json.Unmarshal([]byte(mockKb.String()), &newKb); err != nil {
		t.Error("can't unmarshall pkgs")
	}
	if !(newKb.Packages["available"][0] == ogKb.Packages["available"][0]) {
		t.Error("available pkgs don't match")
	}
	if !(newKb.Packages["available_security"][0] == ogKb.Packages["available_security"][0]) {
		t.Error("available_security pkgs don't match")
	}
	if !(newKb.Packages["installed"][0] == ogKb.Packages["installed"][0]) {
		t.Error("installed pkgs don't match")
	}

}
