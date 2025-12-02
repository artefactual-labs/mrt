#!/usr/bin/env bash

set -e

__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

mkdir -p "$__dir/dist/assets"

generate_runc() {
    local arch=$1
    local version=$2
    echo "Downloading runc.$arch ($version)..."
    curl -Ls "https://github.com/opencontainers/runc/releases/download/$version/runc.$arch" > "$__dir/dist/assets/runc.$arch"
    chmod +x "$__dir/dist/assets/runc.$arch"
}

generate_rootfs() {
    local arch=$1
    local image=$2
    local platform=linux/$arch
    echo "Downloading $image ($platform)..."
    docker pull --quiet --platform=$platform $image 1>/dev/null
    container_id=$(docker create --quiet --platform=$platform $image)
    docker export -o "$__dir/dist/assets/rootfs.$arch.tar" $container_id
    zstd --rm --quiet --force --compress "$__dir/dist/assets/rootfs.$arch.tar"
    md5sum "$__dir/dist/assets/rootfs.$arch.tar.zst" | awk '{ print $1 }' > "$__dir/dist/assets/rootfs.$arch.tar.zst.md5"
    docker rm $container_id 1>/dev/null
}

generate_runc amd64 v1.4.0
generate_runc arm64 v1.4.0
generate_rootfs amd64 python:3.14.0-alpine3.22
generate_rootfs arm64 python:3.14.0-alpine3.22

echo "All assets have been generated successfully!"
