package main

import (
	"fmt"
	"os"
	"reflect"

	"github.com/cleardataeng/mirach/cmd"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			if reflect.TypeOf(r).String() == "plugin.Exception" {
				fmt.Println(r)
				return
			}
			fmt.Println(r)
			os.Exit(1)
		}
	}()
	if err := cmd.MirachCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
