package main

import (
	"log"
	"os"

	"github.com/larashed/agent-go/agent"
)

func main() {
	log.SetOutput(os.Stdout)

	application := agent.NewApp()
	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
