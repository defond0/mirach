package mirachlib

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func readAssetID() string {
	var valid = regexp.MustCompile(`^[A-Za-z0-9]([\w-]*[A-Za-z0-9])?$`)
	var in string
	for valid.MatchString(in) == false {
		fmt.Print("asset id: ")
		stdin := bufio.NewReader(os.Stdin)
		read, _ := stdin.ReadString('\n')
		in = strings.TrimRight(read, "\n")
		if valid.MatchString(in) == false {
			fmt.Println("valid values: starts and ends with alphanumeric, can contain dashes and underscores")
		}
	}
	return in
}
