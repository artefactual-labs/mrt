package dist

import "embed"

//go:embed assets/runc.arm64 assets/rootfs.arm64.tar.zst assets/rootfs.arm64.tar.zst.md5
var assets embed.FS
