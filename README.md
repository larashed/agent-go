Larashed Go Agent
==============

[![Build Status](https://travis-ci.com/larashed/agent-go.svg?branch=master)](https://travis-ci.com/larashed/agent-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/larashed/agent-go)](https://goreportcard.com/report/github.com/larashed/agent-go)

Larashed Go agent starts a socket server and collects metrics from your server and Laravel application.
These metrics are then sent to [larashed.com](https://larashed.com/).

## Collected metrics
- Server load 
- CPU usage
- Memory usage
- Disk space usage
- Operating system name and version
- Boot time
- Whether a reboot is required
- Docker container metrics
- PHP version

## Platform support

We currently support macOS and major Linux (amd64) distributions. Thanks to the nature of Golang, we should be able
 to add more platforms quite easily per your request. Get in touch!

## How to run

### Communication and configuration

Agent collects metrics through TCP or a Unix domain socket. Your application's configuration should match the
 transport method.

### Install a systemd service (recommended)
```
curl -sSL 'https://install.larashed.com/linux' | sudo LARASHED_APP_ID='xxxx' LARASHED_APP_KEY='zzzz' LARASHED_APP_ENV='production' sh
```

The following environment variables will be read if present:
- `LARASHED_APP_ID`
- `LARASHED_APP_KEY`
- `LARASHED_APP_ENV`
- `LARASHED_SOCKET_TYPE`
- `LARASHED_SOCKET_ADDRESS`
- `LARASHED_ARGS`

Agent configuration will be stored in `/etc/larashed/larashed.conf`.

### Update agent

Download the script:

```
curl -sSL 'https://install.larashed.com/linux' -o /tmp/larashed-installer.sh && chmod +x /tmp/larashed-installer.sh
```

#### Update agent to the latest version

```
sudo /tmp/larashed-installer.sh --update
```

#### Completely uninstall the agent

```
sudo /tmp/larashed-installer.sh --uninstall
```

### Manual run

Download the latest binary from the [releases](https://github.com/larashed/agent-go/releases/latest) page and run:

```
agent_linux_amd64 run \
    --app-id=xxxxx \
    --app-key=xxxxx \
    --app-env=production \
    --socket-type=tcp \
    --socket-address=0.0.0.0:33101
```

OPTIONS:
```
--socket-type value            Socket type (unix, tcp) (default: "unix")
--socket-address value         Socket address
--socket value                 Socket address (deprecated, use --socket-address instead)
--api-url value                Larashed API URL (default: "https://api.larashed.com/")
--env value, --app-env value   Application's environment name
--app-id value                 Your application's ID
--app-key value                Your application's secret key
--path-proc value              Kernel & process file path (default: "/proc")
--path-sys value               System component file path (default: "/sys")
--hostname value               Hostname
--log-level value              Logging level (info, debug, trace) (default: "debug")
--collect-server-resources     Collect server resource metrics (default: true)
--collect-application-metrics  Collect application metrics (default: true)
--help, -h                     show help (default: false)
```

### Docker

You can run our agent as a Docker container.

> We recommend you disable container OS resource monitoring using the `--collect-server-resources=false` flag and use
> the agent container to collect application metrics only. **To monitor your container resource usage, install the
> monitoring agent on the host machine.**

To start the latest tagged image, run:
```
docker run -it \
    larashed/agent:latest \
    --app-id=xxxxx \
    --app-key=xxxxx \
    --app-env=production \
    --socket-type=tcp \
    --collect-server-resources=false \
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
    - "--collect-server-resources=false"
    - "--hostname=your_hostname"
```

---
While not recommended, you can monitor basic host machine metrics by mounting your 
`/proc` and `/sys` directories to `/host` container directory.

Docker run:
```
docker run -it \
    ...
    -v /proc:/host/proc:ro \
    -v /sys:/host/sys:ro
```

Docker compose:
```
  volumes:
    - "/proc:/host/proc:ro"
    - "/sys:/host/sys:ro"
```