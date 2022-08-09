package main

import (
	"fmt"
	"os"
)

func main() {
	cd, _ := os.Getwd()
	exe, _ := os.Executable()
	fmt.Printf("---Golang script---\n")
	fmt.Printf(" + os.Args: %#v\n", os.Args)
	fmt.Printf(" + working directory: %s\n", cd)
	fmt.Printf(" + application path: %s\n", exe)
}
