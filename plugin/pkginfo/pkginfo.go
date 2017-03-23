package pkginfo

import (
	"encoding/json"

	"gitlab.eng.cleardata.com/dash/mirach/plugin"
	"gitlab.eng.cleardata.com/dash/mirach/plugin/pkginfo/parsers"

	"github.com/shirou/gopsutil/host"
)

// PkgStatus represents the OS and map of list of LinuxPackage.
type PkgStatus struct {
	OS       string
	Packages map[string][]parsers.LinuxPackage `json:"pkg_info"`
}

// KBStatus represents the OS and map of list of KBArticle.
type KBStatus struct {
	Articles map[string][]parsers.KBArticle `json:"pkg_info"`
}

//GetInfo fill in the package status object with info.
func (p *PkgStatus) GetInfo() {
	switch p.OS {
	case "debian":
		packages, _ := parsers.GetAptDpkgPkgs()
		p.Packages = packages
	case "rhel":
		packages, _ := parsers.GetYumPkgs()
		p.Packages = packages
	}
}

//String returns the filled in data from PkgStatus as str.
func (p *PkgStatus) String() string {
	s, _ := json.Marshal(p)
	return string(s)
}

//GetInfoGroup returns the filled in data for given group.
func (p *PkgStatus) GetInfoGroup(infoGroup string) string {
	s, _ := json.MarshalIndent(p.Packages[infoGroup], "", "  ")
	return string(s)
}

//GetInfo fill in the kb status object with info.
func (k *KBStatus) GetInfo() {
	articles, _ := parsers.GetWindowsKBs()
	k.Articles = articles
}

//GetInfoGroup returns the filled in data for given group.
func (k *KBStatus) GetInfoGroup(infoGroup string) string {
	s, _ := json.MarshalIndent(k.Articles[infoGroup], "", "  ")
	return string(s)
}

//String returns the filled in data from PkgStatus as str.
func (k *KBStatus) String() string {
	s, _ := json.MarshalIndent(k, "", "  ")
	return string(s)
}

//GetInfo will load up and return InfoGroup for current OS.
func GetInfo() plugin.InfoGroup {
	os := getOS()
	if os == "windows" {
		kb := new(KBStatus)
		kb.GetInfo()
		return kb
	}
	pkg := new(PkgStatus)
	pkg.OS = os
	pkg.GetInfo()
	return pkg
}

//String will load up and return InfoGroup for current OS.
func String() string {
	os := getOS()
	if os == "windows" {
		kb := new(KBStatus)
		kb.GetInfo()
		return kb.String()
	}
	pkg := new(PkgStatus)
	pkg.OS = os
	pkg.GetInfo()
	return pkg.String()

}

//GetInfoGroup will load up and return InfoGroup for current OS.
func GetInfoGroup(infoGroup string) string {
	os := getOS()
	if os == "windows" {
		kb := new(KBStatus)
		kb.GetInfo()
		return kb.GetInfoGroup(infoGroup)
	}
	pkg := new(PkgStatus)
	pkg.OS = os
	pkg.GetInfo()
	return pkg.GetInfoGroup(infoGroup)
}

func getOS() string {
	host, err := host.Info()
	if err != nil {
		panic(err)
	}
	if host.OS != "windows" {
		return host.PlatformFamily
	}
	return host.OS
}
