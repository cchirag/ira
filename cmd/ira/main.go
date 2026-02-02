package main

import (
	"embed"
	"fmt"

	"github.com/cchirag/ira/internal/spawn"
)

//go:embed bin/*
var binaryFS embed.FS

func main() {
	if err := spawn.RunDaemon(binaryFS); err != nil {
		fmt.Println("error strting daemon: ", err.Error())
	}
	fmt.Println("Welcome to Ira!!")
}
