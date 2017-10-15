package input

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/theherk/viper"
)

// ReadBroker gets a broker URL from input.
func ReadBroker() string {
	var in string
	for {
		fmt.Print("mqtt broker url: ")
		stdin := bufio.NewReader(os.Stdin)
		read, _ := stdin.ReadString('\n')
		in = strings.TrimRight(read, "\n")
		if _, err := url.ParseRequestURI(in); err != nil {
			fmt.Printf("broker url not valid: %s; try again\n", err.Error())
			continue
		}
		return in
	}
}

// ReadCfgType gets the preferred configuration type from input.
func ReadCfgType() string {
	fmt.Println("creating blank configuration file")
	var in string
	for {
		fmt.Printf("config type, %s (default: yaml): ", viper.SupportedExts)
		stdin := bufio.NewReader(os.Stdin)
		read, _ := stdin.ReadString('\n')
		in = strings.TrimRight(read, "\n")
		if in == "" {
			return "yaml"
		}
		for _, t := range viper.SupportedExts {
			if in == t {
				return t
			}
		}
		fmt.Println("must leave blank or enter one of the valid values given above; try again")
	}
}

// ReadID get the machine's ID from input.
func ReadID() string {
	var valid = regexp.MustCompile(`^[A-Za-z0-9]([\w-]*[A-Za-z0-9])?$`)
	var in string
	for valid.MatchString(in) == false {
		fmt.Print("machine id: ")
		stdin := bufio.NewReader(os.Stdin)
		read, _ := stdin.ReadString('\n')
		in = strings.TrimRight(read, "\n")
		if valid.MatchString(in) == false {
			fmt.Println("valid values: starts and ends with alphanumeric, can contain dashes and underscores")
		}
	}
	return in
}
