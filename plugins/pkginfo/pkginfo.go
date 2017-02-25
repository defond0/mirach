package pkginfo

import (
	"encoding/json"

	"cleardata.com/dash/mirach/plugins/pkginfo/parsers"
	"github.com/shirou/gopsutil/host"
)

// InfoGroup is an interface for getting data and marshaling to json.
type InfoGroup interface {
	GetInfo()
	String() string
}

type PkgStatus struct {
	Os       string
	Packages map[string][]parsers.LinuxPackage `json:"packages"`
}

type KBStatus struct {
	Articles map[string][]parsers.KBArticle `json:"articles"`
}

func (p *PkgStatus) GetInfo() {
	switch p.Os {
	case "debian":
		packages, _ := parsers.GetAptDpkgPkgs()
		p.Packages = packages
	case "rhel":
		packages, _ := parsers.GetYumPkgs()
		p.Packages = packages
	}
}

func (p *PkgStatus) String() string {
	s, _ := json.Marshal(p)
	return string(s)
}

func (p *PkgStatus) GetInfoGroup(infoGroup string) string {
	s, _ := json.MarshalIndent(p.Packages[infoGroup], "", "  ")
	return string(s)
}

func (k *KBStatus) GetInfo() {
	articles, _ := parsers.GetWindowsKBs()
	k.Articles = articles
}

func (k *KBStatus) GetInfoGroup(infoGroup string) string {
	s, _ := json.MarshalIndent(k.Articles[infoGroup], "", "  ")
	return string(s)
}
func (k *KBStatus) String() string {
	s, _ := json.MarshalIndent(k, "", "  ")
	return string(s)
}

func GetInfo() {
	os := getOs()
	if os == "windows" {
		kb := new(KBStatus)
		kb.GetInfo()

	} else {
		pkg := new(PkgStatus)
		pkg.Os = os
		pkg.GetInfo()
	}

}

func String() string {
	os := getOs()
	if os == "windows" {
		kb := new(KBStatus)
		kb.GetInfo()
		return kb.String()

	} else {
		pkg := new(PkgStatus)
		pkg.Os = os
		pkg.GetInfo()
		return pkg.String()
	}

}

func GetInfoGroup(infoGroup string) string {
	os := getOs()
	if os == "windows" {
		kb := new(KBStatus)
		kb.GetInfo()
		return kb.GetInfoGroup(infoGroup)

	} else {
		pkg := new(PkgStatus)
		pkg.Os = os
		pkg.GetInfo()
		return pkg.GetInfoGroup(infoGroup)
	}
}

func getOs() string {
	host, err := host.Info()
	if err != nil {
		panic(err)
	}
	if host.OS != "windows" {
		return host.PlatformFamily
	} else {
		return host.OS
	}
}
