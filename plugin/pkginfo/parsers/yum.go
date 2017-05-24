package parsers

import "os/exec"

// GetYumPkgs creates map of available, installed and available security packages from yum as well as a list of errors that occurred generating that list. grep returns exit status 1 when it gets no match, errors like that are fine
func GetYumPkgs() (map[string]map[string]LinuxPackage, []error) {
	errors := []error{}
	out := make(map[string]map[string]LinuxPackage)
	avail, err := getYumAvailablePackages()
	if err != nil {
		errors = append(errors, err)
	}
	out["available"] = avail
	availSec, err := getYumAvailableSecurityPackages()
	if err != nil {
		errors = append(errors, err)
	}
	out["available_security"] = availSec
	installed, err := getYumInstalledPackages()
	if err != nil {
		errors = append(errors, err)
	}
	out["installed"] = installed
	return out, errors
}

func getYumInstalledPackages() (map[string]LinuxPackage, error) {
	yum := command("yum list installed -q")
	grep := exec.Command("grep", "-v", "Installed Packages")
	awk := exec.Command("awk", "{{ print $1 , $2 }}")
	stdout, _, err := pipeline(yum, grep, awk)
	if err != nil {
		return nil, err
	}
	return parsePacakgesFromBytes(stdout, false)
}

func getYumAvailablePackages() (map[string]LinuxPackage, error) {
	yum := command("yum list updates -q")
	grep := exec.Command("grep", "-v", "Updated Packages")
	awk := exec.Command("awk", "{{ print $1 , $2 }}")
	stdout, _, err := pipeline(yum, grep, awk)
	if err != nil {
		return nil, err
	}
	return parsePacakgesFromBytes(stdout, false)
}

func getYumAvailableSecurityPackages() (map[string]LinuxPackage, error) {
	yum := command("yum list updates -q --security")
	grep := exec.Command("grep", "-v", "Updated Packages")
	awk := exec.Command("awk", "{{ print $1 , $2 }}")
	stdout, _, err := pipeline(yum, grep, awk)
	if err != nil {
		return nil, err
	}
	return parsePacakgesFromBytes(stdout, true)
}
