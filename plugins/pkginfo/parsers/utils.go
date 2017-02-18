package parsers

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func command(cmd string) *exec.Cmd {
	cmd_slice := strings.Split(cmd, " ")
	return exec.Command(cmd_slice[0], cmd_slice[1:]...)
}

//Credit where credit is due, this function is almost entirely from:
//https://gist.github.com/kylelemons/1525278
func pipeline(cmds ...*exec.Cmd) (pipeLineOutput, collectedStandardError []byte, pipeLineError error) {
	// Require at least one command
	if len(cmds) < 1 {
		return nil, nil, nil
	}
	// Collect the output from the command(s)
	var output bytes.Buffer
	var stderr bytes.Buffer
	last := len(cmds) - 1
	for i, cmd := range cmds[:last] {
		var err error
		// Connect each command's stdin to the previous command's stdout
		if cmds[i+1].Stdin, err = cmd.StdoutPipe(); err != nil {
			return nil, nil, err
		}
		// Connect each command's stderr to a buffer
		cmd.Stderr = &stderr
	}
	// Connect the output and error for the last command
	cmds[last].Stdout, cmds[last].Stderr = &output, &stderr
	// Start each command
	for _, cmd := range cmds {
		if err := cmd.Start(); err != nil {
			return output.Bytes(), stderr.Bytes(), err
		}
	}
	// Wait for each command to complete
	for _, cmd := range cmds {
		if err := cmd.Wait(); err != nil {
			return output.Bytes(), stderr.Bytes(), err
		}
	}
	// Return the pipeline output and the collected standard error
	return output.Bytes(), stderr.Bytes(), nil
}

func getAptitudeSecurityList() {
	cmd := command("grep security /etc/apt/sources.list")
	outfile, err := os.Create("/tmp/security.list")
	if err != nil {
		fmt.Println(err)
	}
	defer outfile.Close()
	cmd.Stdout = outfile

	err = cmd.Start()
	if err != nil {
		fmt.Println(err)
	}

	err = cmd.Wait()
	if err != nil {
		fmt.Println(err)
	}
}

func cleanUpSecurityList() {
	cmd := command("rm /tmp/security.list")
	err := cmd.Start()
	if err != nil {
		fmt.Println(err)
	}

	err = cmd.Wait()
	if err != nil {
		fmt.Println(err)
	}
}

//Expects []bytes representing lines of pkgname version
func parsePacakgesFromBytes(b []byte, security bool) ([]LinuxPackage, error) {
	pkgs := []LinuxPackage{}
	name, _ := regexp.Compile("[^\\s]+")
	version, _ := regexp.Compile("\\s(.*)")
	lines := strings.Split(string(b), "\n")
	for _, line := range lines {
		if name.MatchString(line) {
			pkgs = append(pkgs,
				LinuxPackage{
					Name:     name.FindString(line),
					Version:  strings.Trim(version.FindString(line), "[ ]"),
					Security: security,
				},
			)
		}
	}
	return pkgs, nil
}

//Expects []bytes representing lines of pkgname version
func parseArticlesFromBytes(b []byte, security bool) ([]KBArticle, error) {
	art := []KBArticle{}
	name, _ := regexp.Compile("[^\\s]+")
	lines := strings.Split(string(b), "\n")
	for _, line := range lines {
		if name.MatchString(line) {
			art = append(art,
				KBArticle{
					Name:     name.FindString(line),
					Security: security,
				},
			)
		}
	}
	return art, nil
}
