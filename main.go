package main

import (
	"fmt"
	"os"

	"github.com/cleardataeng/mirach/cmd"
)

//go:generate go run gen.go

func main() {
	if err := cmd.MirachCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
