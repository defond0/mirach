package parsers

import (
	"os"
	"os/exec"
)

// GetAptDpkgPkgs creates map of available, installed and available security packages from aptitude and dpkg as well as a list of errors that occurred generating that list. grep returns exit status 1 when it gets no match, errors like that are fine
func GetAptDpkgPkgs() (map[string]map[string]LinuxPackage, []error) {
	errors := []error{}
	out := make(map[string]map[string]LinuxPackage)
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

func getDpkgInstalledPackages() (map[string]LinuxPackage, error) {
	aptget := command("dpkg -l")
	sed := exec.Command("sed", "1,/^+++/d")
	awk := exec.Command("awk", "{{ print $2 , $3 }}")
	stdout, _, err := pipeline(aptget, sed, awk)
	if err != nil {
		return nil, err
	}
	return parsePacakgesFromBytes(stdout, false)
}

func getAptAvailablePackages() (map[string]LinuxPackage, error) {
	aptget := command("apt-get upgrade -qq --just-print")
	grep := command("grep Inst")
	awk := exec.Command("awk", "{{ print $2 , $3 }}")

	stdout, _, err := pipeline(aptget, grep, awk)
	if err != nil {
		return nil, err
	}
	return parsePacakgesFromBytes(stdout, false)
}

func getAptAvailableSecurityPackages() (map[string]LinuxPackage, error) {
	aptget := command("apt-get upgrade -oDir::Etc::Sourcelist=/tmp/security.list -oDir::Etc::Sourceparts='-' -oDir::Etc::Vendorlist='-' -oDir::Etc::Vendorparts='-' -qq --just-print")
	grep := command("grep Inst")
	awk := exec.Command("awk", "{{ print $2 , $3 }}")
	stdout, _, err := pipeline(aptget, grep, awk)
	if err != nil {
		return nil, err
	}
	return parsePacakgesFromBytes(stdout, true)
}

func getAptitudeSecurityList() error {
	cmd := command("grep security /etc/apt/sources.list")
	outfile, err := os.Create("/tmp/security.list")
	if err != nil {
		return err
	}
	defer outfile.Close()
	cmd.Stdout = outfile
	err = cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}

func cleanUpSecurityList() error {
	cmd := command("rm /tmp/security.list")
	err := cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}
