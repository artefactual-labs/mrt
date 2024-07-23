# mrt

mrt (My Runtime) is a proof of concept where a Go application runs a rootless
container on Linux without additional dependencies. To achieve this, mrt embeds
[runc] and an [OCI bundle] based on [python:3.12.4-alpine3.20].

## Hypothesis

In the context of Archivematica, we hypothesize that it is possible to implement
a Go application capable of executing a pool of workers (MCPClient) using
rootless containers without additional dependencies such as Docker or Podman.
This implementation can be achieved by embedding runc and dynamically generating
the OCI bundle from both an embedded root filesystem and by pulling images from
published repositories on demand.

This initiative emerges within the CCP project, where we are exploring
alternative distributions of Archivematica, focusing on self-contained solutions
for single-node environments. In this context, software dependencies typically
distributed along with the `archivematica-mcp-client` deb or rpm packages could
instead be bundled as part of the application binary.

## Supported architectures

These are the platform-architecture combinations for which mrt provides
compatibility:

* linux-amd64
* linux-arm64

## Usage

Binaries are not published since this project is a proof of concept. Build the
binary from sources as follows:

```
$ go build -o /tmp/mrt .; /tmp/mrt
2024-07-18T13:31:10.106+0200	V(0)	mrt	mrt/main.go:170	Using cached rootfs.
2024-07-18T13:31:10.120+0200	V(0)	mrt	mrt/main.go:63	Container deleted.	{"id": "arbutus"}
2024-07-18T13:31:10.120+0200	V(0)	mrt	mrt/main.go:72	Creating container	{"id": "arbutus"}
Python 3.12.4
2024-07-18T13:31:10.185+0200	V(0)	mrt	mrt/main.go:79	Container executed!	{"pid": 0}
```

It runs `python -V` inside the container and print the output.

## Generate assets

Use `generate.sh` to generate all the assets required:

```
$ ./generate.sh
Downloading runc.arm64...
Downloading runc.amd64...
Downloading python:3.12.4-alpine3.20 (linux/amd64)...
Downloading python:3.12.4-alpine3.20 (linux/arm64)...
All assets have been generated successfully!
```


[runc]: https://github.com/opencontainers/runc
[python:3.12.4-alpine3.20]: https://hub.docker.com/layers/library/python/3.12.4-alpine3.20/images/sha256-ebe4166fcf7fd212975cb932440ba69cfd6c27fdb9ab2253f965a1d2d7f1c476
[OCI bundle]: https://github.com/opencontainers/runtime-spec/blob/main/bundle.md
