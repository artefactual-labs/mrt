package dist

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
)

var (
	runcPath      = fmt.Sprintf("assets/runc.%s", runtime.GOARCH)
	rootFSPAth    = fmt.Sprintf("assets/rootfs.%s.tar.zst", runtime.GOARCH)
	rootFSSumPath = fmt.Sprintf("assets/rootfs.%s.tar.zst.md5", runtime.GOARCH)
)

func WriteRunc(dest string) error {
	return Write(runcPath, dest, os.FileMode(0o750))
}

func WriteRootFS(dest string) error {
	return Write(rootFSPAth, dest, os.FileMode(0o640))
}

func WriteRootFSSum(dest string) error {
	return Write(rootFSSumPath, dest, os.FileMode(0o640))
}

func Write(path, dest string, mode os.FileMode) error {
	src, err := assets.Open(path)
	if err != nil {
		return err
	}

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if err := os.Chmod(dest, mode); err != nil {
		return err
	}

	_, err = io.Copy(destFile, src)
	if err != nil {
		return err
	}

	err = destFile.Sync()
	if err != nil {
		return err
	}

	return nil
}

func CheckRootFSSum(sum []byte) bool {
	content, err := assets.ReadFile(rootFSSumPath)
	if err != nil {
		return false
	}

	return bytes.Equal(content, sum)
}
