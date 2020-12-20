package collectors

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestDocker(t *testing.T) {
	cl, err := NewDockerClient()
	if err != nil {
		panic(err)
	}
	containers, err := cl.FetchContainers()
	if err != nil {
		panic(err)
	}
	spew.Dump(containers)

	containers, err = cl.FetchContainers()
	if err != nil {
		panic(err)
	}
	spew.Dump(containers)
}