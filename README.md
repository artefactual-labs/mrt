# mrt

mrt (My Runtime) is a proof of concept where a Go application runs a rootless
container on Linux without additional dependencies. To achieve this, mrt embeds
[runc] and an OCI bundle based on [docker.io/library/python:3.12.4-alpine3.20].

## Usage

```
$ go build -o /tmp/mrt .
$ /tmp/mrt
2024/07/17 21:27:04 Container deleted.
2024/07/17 21:27:04 Creating container...
2024/07/17 21:27:07 Container executed - pid 0
```

The container runs for three seconds because it executes `sleep 3`.

## Assets

### `dist/assets/runc`

Downloaded from https://github.com/opencontainers/runc/releases.

### `dist/assets/rootfs.tar`

```
docker export -o rootfs (docker create python:3.12.4-alpine3.20)
```


[runc]: https://github.com/opencontainers/runc
[docker.io/library/python:3.12.4-alpine3.20]: https://hub.docker.com/layers/library/python/3.12.4-alpine3.20/images/sha256-ebe4166fcf7fd212975cb932440ba69cfd6c27fdb9ab2253f965a1d2d7f1c476
