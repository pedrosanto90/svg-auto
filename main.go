package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Welcome SVG Auto!")
	
	fmt.Println(getFiles())
}

func getFiles() []string {
	var args []string
	
	for _, arg := range os.Args[1:] {
		args = append(args, arg)
	}
	return args
}

func accessIcoMoon() {
	// this function should access icomoon.io
}

func uploadFiles() {
	// this function should updload files to icommon.io
}

func downloadFiles() {
	// this function should download files from iconmoon.io
}
