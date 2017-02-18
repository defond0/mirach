package parsers

import (
	"fmt"
	"os/exec"
)

//grep returns exit status 1 when it gets no match, errors like that are fine
func GetAptDpkgPkgs() (map[string][]LinuxPackage, []error) {
	errors := []error{}
	out := make(map[string][]LinuxPackage)
	avail, err := getAptAvailablePackages()
	if err != nil {
		errors = append(errors, err)
	}
	out["available"] = avail
	getAptitudeSecurityList()
	defer cleanUpSecurityList()
	avail_sec, err := getAptAvailableSecurityPackages()
	if err != nil {
		errors = append(errors, err)
	}
	out["available_security"] = avail_sec
	installed, err := getDpkgInstalledPackages()
	if err != nil {
		errors = append(errors, err)
	}
	out["installed"] = installed
	return out, errors
}

func getDpkgInstalledPackages() ([]LinuxPackage, error) {
	aptget := command("dpkg -l")
	awk := exec.Command("awk", "{{ print $2 , $3 }}")

	stdout, stderr, err := pipeline(aptget, awk)
	if err != nil {
		fmt.Println(string(stderr))
		return nil, err
	}
	return parsePacakgesFromBytes(stdout, false)
}

func getAptAvailablePackages() ([]LinuxPackage, error) {
	aptget := command("apt-get upgrade -qq --just-print")
	grep := command("grep Inst")
	awk := exec.Command("awk", "{{ print $2 , $3 }}")

	stdout, stderr, err := pipeline(aptget, grep, awk)
	if err != nil {
		fmt.Println(stderr)
		return nil, err
	}
	return parsePacakgesFromBytes(stdout, false)
}

func getAptAvailableSecurityPackages() ([]LinuxPackage, error) {
	aptget := command("apt-get upgrade -oDir::Etc::Sourcelist=/tmp/security.list -oDir::Etc::Sourceparts='-' -oDir::Etc::Vendorlist='-' -oDir::Etc::Vendorparts='-' -qq --just-print")
	grep := command("grep Inst")
	awk := exec.Command("awk", "{{ print $2 , $3 }}")
	stdout, stderr, err := pipeline(aptget, grep, awk)
	if err != nil {
		fmt.Println(string(stderr))
		return nil, err
	}
	return parsePacakgesFromBytes(stdout, true)
}
