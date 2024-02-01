package main

import (
	"log"
	"os"
	"os/exec"
)

func main() {
	homePath := os.Getenv("HOME")
	cmd := exec.Command(homePath+"/go/bin/swag", "init")

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}
