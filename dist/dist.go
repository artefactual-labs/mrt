package dist

import (
	"bytes"
	"embed"
	"io"
	"os"
)

//go:embed assets/*
var Assets embed.FS

func MatchRootFSChecksum(sum []byte) bool {
	content, err := Assets.ReadFile("assets/rootfs.tar.zst.md5")
	if err != nil {
		return false
	}

	return bytes.Equal(content, sum)
}

func Write(path, dest string, mode os.FileMode) error {
	src, err := Assets.Open(path)
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
