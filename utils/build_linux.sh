export GIT_COMMIT=$(git rev-list -1 HEAD)

env GOOS=linux \
    GOARCH=amd64 \
    go build -ldflags "
      -X 'github.com/larashed/agent-go/config.GitCommit=$GIT_COMMIT'
      -X 'github.com/larashed/agent-go/config.GitTag=$TRAVIS_TAG'
    " -o \
    build/linux/agent .