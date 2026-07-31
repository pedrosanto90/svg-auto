package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Welcome SVG Auto!")
	getFiles()
}

func getFiles() {
	args := os.Args
	fmt.Println("Args:", args[1:])
}
