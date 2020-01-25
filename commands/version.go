package commands

import (
	"encoding/json"
	"fmt"

	"github.com/larashed/agent-go/config"
)

func NewVersionCommand(isJson bool) {
	v := config.Version{
		Tag:    config.GitTag,
		Commit: config.GitCommit,
	}

	if isJson {
		j, _ := json.Marshal(v)
		fmt.Println(string(j))
		return
	}

	fmt.Println("Version:", v.Tag)
	fmt.Println("Commit:", v.Commit)
}
