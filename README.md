# mrt

mrt (My Runtime) is a proof of concept where a Go application runs a rootless
container on Linux without additional dependencies. To achieve this, mrt embeds
[runc] and an [OCI bundle] based on a Python image using a modern release of
Python 3.

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
binary from the sources as follows:

    go build -o /tmp/mrt .

Execute the binary:

    $ /tmp/mrt
    2025-01-14T12:21:38.761+0100    V(0)    mrt     mrt/main.go:183 Using cached rootfs.
    2025-01-14T12:21:38.765+0100    V(0)    mrt     mrt/main.go:64  Using runc.     {"version": "1.2.4", "path": "/home/jesus/.cache/mrt/runc"}
    2025-01-14T12:21:38.768+0100    V(0)    mrt     mrt/main.go:70  Container deleted.      {"id": "arbutus"}
    2025-01-14T12:21:38.768+0100    V(0)    mrt     mrt/main.go:79  Creating container      {"id": "arbutus"}
    Python 3.14.0 (x86_64)
    2025-01-14T12:21:38.919+0100    V(0)    mrt     mrt/main.go:86  Container executed!     {"pid": 0}

This is the command we are running inside the container:

    python -c "import platform; print(f'Python {platform.python_version()} ({platform.machine()})')"

## Rootless containers

`mrt` automatically attempts to create containers in rootless mode if it detects
that the user running `mrt` does not have root privileges. However, this mode
introduces some additional complexities, as outlined below.

### Enabling user namespaces

User namespaces must be enabled on the system. Run the following command:

    sudo sysctl -w kernel.unprivileged_userns_clone=1

To make this change persistent, add the following line to `/etc/sysctl.conf`:

    kernel.unprivileged_userns_clone=1

### Configuring security frameworks

In environments using AppArmor, the `runc` binary deployed by `mrt` may require
a specific profile to execute. For example, adjusting the path as needed, add
the following profile to `/etc/apparmor.d/mrt`:

```plaintext
profile mrt /home/user/.cache/mrt/runc flags=(unconfined) {
    userns,
    /home/user/.cache/mrt/runc ix,
}
```

Load the profile with:

    sudo apparmor_parser -r /etc/apparmor.d/mrt

## Generate assets

Use `generate.sh` to generate all the assets required:

```
$ ./generate.sh
Downloading runc.arm64...
Downloading runc.amd64...
Downloading python:3.14.0-alpine3.22 (linux/amd64)...
Downloading python:3.14.0-alpine3.22 (linux/arm64)...
All assets have been generated successfully!
```


[runc]: https://github.com/opencontainers/runc
[OCI bundle]: https://github.com/opencontainers/runtime-spec/blob/main/bundle.md
