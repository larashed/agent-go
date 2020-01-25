package config

type Version struct {
	Tag    string `json:"tag"`
	Commit string `json:"commit"`
}

var (
	GitTag    string
	GitCommit string
)
