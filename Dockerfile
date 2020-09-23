FROM golang:1.15-alpine AS builder
LABEL stage=builder

RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

WORKDIR $GOPATH/src/larashed/agent-go

COPY . .

RUN go get -d -v

ARG SOURCE_COMMIT
ENV SOURCE_COMMIT $SOURCE_COMMIT

ARG SOURCE_BRANCH
ENV SOURCE_BRANCH $SOURCE_BRANCH

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
   go build -ldflags " \
     -X 'github.com/larashed/agent-go/config.GitCommit=$SOURCE_COMMIT' \
     -X 'github.com/larashed/agent-go/config.GitTag=$SOURCE_BRANCH' \
   " -o \
   /go/bin/agent .

FROM scratch

ENV DOCKER_BUILD=1

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/agent /go/bin/agent

VOLUME /host/proc /host/sys

ENTRYPOINT ["/go/bin/agent", "run"]