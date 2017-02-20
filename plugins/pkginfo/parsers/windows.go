package parsers

import (
	"fmt"
	"os/exec"
)

//grep returns exit status 1 when it gets no match, errors like that are fine
func GetWindowsKBs() (map[string][]KBArticle, []error) {
	errors := []error{}
	out := make(map[string][]KBArticle)
	avail, err := getYumAvailableKBs()
	if err != nil {
		errors = append(errors, err)
	}
	out["available"] = avail
	avail_sec, err := getYumAvailableSecurityKBs()
	if err != nil {
		errors = append(errors, err)
	}
	out["available_security"] = avail_sec
	installed, err := getWindowsInstalledKBs()
	if err != nil {
		errors = append(errors, err)
	}
	out["installed"] = installed
	return out, errors
}

func getWindowsInstalledKBs() ([]KBArticle, error) {
	wmicList := command("wmic qfe list")
	stdout, stderr, err := pipeline(wmicList)
	if err != nil {
		fmt.Println(string(stderr))
		return nil, err
	}
	return parseArticlesFromBytes(stdout)
}

func getYumAvailableKBs() ([]KBArticle, error) {
	aptget := command("yum list updates -q")
	grep := exec.Command("grep", "-v", "Updated KBs")
	awk := exec.Command("awk", "{{ print $1 , $2 }}")
	stdout, stderr, err := pipeline(aptget, grep, awk)
	if err != nil {
		fmt.Println(stderr)
		return nil, err
	}
	return parseArticlesFromBytes(stdout)
}

func getYumAvailableSecurityKBs() ([]KBArticle, error) {
	aptget := command("yum list updates -q --security")
	grep := exec.Command("grep", "-v", "Updated KBs")
	awk := exec.Command("awk", "{{ print $1 , $2 }}")
	stdout, stderr, err := pipeline(aptget, grep, awk)
	if err != nil {
		fmt.Println(string(stderr))
		return nil, err
	}
	return parseArticlesFromBytes(stdout)
}
