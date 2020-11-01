package config

// Version holds build tag and commit
// these are injected during build
type Version struct {
	Tag    string `json:"tag"`
	Commit string `json:"commit"`
}

var (
	// GitTag holds git tag
	GitTag string
	// GitCommit holds git commit hash
	GitCommit string
)
