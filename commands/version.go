package commands

import (
	"encoding/json"
	"fmt"

	"github.com/larashed/agent-go/config"
)

// NewVersionCommand prints agent version
func NewVersionCommand(isJSON bool) {
	v := config.Version{
		Tag:    config.GitTag,
		Commit: config.GitCommit,
	}

	if isJSON {
		j, _ := json.Marshal(v)
		fmt.Println(string(j))
		return
	}

	fmt.Println("Version:", v.Tag)
	fmt.Println("Commit:", v.Commit)
}
