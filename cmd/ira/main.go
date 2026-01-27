package main

import (
	"embed"
	"fmt"
)

//go:embed bin/*
var binaryFS embed.FS

func main() {
	entries, err := binaryFS.ReadDir("bin")
	if err != nil {
		fmt.Printf("error reading binary fs: %s", err.Error())
		return
	}
	for _, entry := range entries {
		fmt.Println("entry: ", entry.Name())
	}
	fmt.Println("Hello World")
}
