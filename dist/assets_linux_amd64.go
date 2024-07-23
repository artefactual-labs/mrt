package dist

import "embed"

//go:embed assets/runc.amd64 assets/rootfs.amd64.tar.zst assets/rootfs.amd64.tar.zst.md5
var assets embed.FS
