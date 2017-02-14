package util

import "fmt"

var version = "undefined"

// ShowVersion will print the version information.
func ShowVersion() {
	fmt.Println("mirach " + version)
}
