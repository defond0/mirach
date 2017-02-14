package main

import (
	"fmt"
	"os"

	"cleardata.com/dash/mirach/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
