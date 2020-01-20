package main

import (
	"log"
	"os"
)

func main() {
	log.SetOutput(os.Stdout)

	application := NewApp()
	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
