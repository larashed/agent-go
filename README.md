Larashed Go Agent
==============

[![Build Status](https://travis-ci.com/larashed/agent-go.svg?branch=master)](https://travis-ci.com/larashed/agent-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/larashed/agent-go)](https://goreportcard.com/report/github.com/larashed/agent-go)

Larashed Go agent starts a socket server and collects metrics from your server and Laravel application.
These metrics are then sent to [larashed.com](https://larashed.com/).

## IPC Communication

There are 2 supported communication types: TCP and Unix domain socket.

**Note:** Make sure program arguments match `larashed/agent` composer package configuration.

## Platform support

We currently support macOS and major Linux (amd64) distributions. Thanks to the nature of Golang, we should be able
 to add more platforms quite easily per your request. Get in touch!

## How to run

Download the latest binary from the [releases](https://github.com/larashed/agent-go/releases/latest) page and run:
```
agent_linux_amd64 run \
    --app-id=xxxxx \
    --app-key=xxxxx \
    --app-env=production \
    --socket-type=tcp \
    --socket-address=0.0.0.0:33101
```
### Docker

You can run our agent as a Docker container.
Check out our images on [Docker hub](https://hub.docker.com/r/larashed/agent/tags).

**Note:** you will have to mount host `/proc` and `/sys` directories to `/host` container directory for server monitoring to
 work.

To start the latest tagged image, run:
```
docker run -it -v /proc:/host/proc:ro \
    -v /sys:/host/sys:ro \
    larashed/agent:latest \
    --app-id=xxxxx \
    --app-key=xxxxx \
    --app-env=production \
    --socket-type=tcp \
    --socket-address=0.0.0.0:33101 \
    --hostname=`hostname`
```

### Docker compose

Identical example using `docker-compose`:

```
agent:
  image: "larashed/agent:latest"
  container_name: agent
  command:
    - "--app-id=xxxxx"
    - "--app-key=xxxxx"
    - "--app-env=production"
    - "--socket-type=tcp"
    - "--socket-address=0.0.0.0:33101"
    - "--hostname=your_hostname"
  volumes:
    - "/proc:/host/proc:ro"
    - "/sys:/host/sys:ro"
```
