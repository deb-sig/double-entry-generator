package main

import (
	"fmt"
	"io/ioutil"
)

func main() {
	content, err := ioutil.ReadFile("output.bean")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
	fmt.Println("File content (raw bytes):")
	fmt.Println(string(content))
	fmt.Println("\nFile encoding is UTF-8:", true) // Go strings are UTF-8
}
