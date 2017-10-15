package main

import (
	"fmt"
	"os"

	"github.com/cleardataeng/mirach/cmd"
)

func main() {
	if err := cmd.MirachCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
