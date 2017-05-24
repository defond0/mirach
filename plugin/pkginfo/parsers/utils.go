package parsers

import (
	"bytes"
	"os/exec"
	"regexp"
	"strings"
)

func command(cmd string) *exec.Cmd {
	cmdSlice := strings.Split(cmd, " ")
	return exec.Command(cmdSlice[0], cmdSlice[1:]...)
}

// Credit where credit is due, this function is almost entirely from:
// https://gist.github.com/kylelemons/1525278
func pipeline(cmds ...*exec.Cmd) (pipeLineOutput, collectedStandardError []byte, pipeLineError error) {
	// Require at least one command.
	if len(cmds) < 1 {
		return nil, nil, nil
	}
	// Collect the output from the command(s).
	var output bytes.Buffer
	var stderr bytes.Buffer
	last := len(cmds) - 1
	for i, cmd := range cmds[:last] {
		var err error
		// Connect each command's stdin to the previous command's stdout.
		if cmds[i+1].Stdin, err = cmd.StdoutPipe(); err != nil {
			return nil, nil, err
		}
		// Connect each command's stderr to a buffer.
		cmd.Stderr = &stderr
	}
	// Connect the output and error for the last command.
	cmds[last].Stdout, cmds[last].Stderr = &output, &stderr
	// Start each command
	for _, cmd := range cmds {
		if err := cmd.Start(); err != nil {
			return output.Bytes(), stderr.Bytes(), err
		}
	}
	// Wait for each command to complete.
	for _, cmd := range cmds {
		if err := cmd.Wait(); err != nil {
			return output.Bytes(), stderr.Bytes(), err
		}
	}
	// Return the pipeline output and the collected standard error.
	return output.Bytes(), stderr.Bytes(), nil
}

// Expects []bytes representing lines of pkgname version.
func parsePacakgesFromBytes(b []byte, security bool) (map[string]LinuxPackage, error) {
	pkgs := map[string]LinuxPackage{}
	name, _ := regexp.Compile("[^\\s]+")
	version, _ := regexp.Compile("\\s(.*)")
	lines := strings.Split(string(b), "\n")
	for _, line := range lines {
		if name.MatchString(line) {
			pkgs[name.FindString(line)] = LinuxPackage{
				name:     name.FindString(line),
				Version:  strings.Trim(version.FindString(line), "[ ]"),
				Security: security,
			}
		}
	}
	return pkgs, nil
}
