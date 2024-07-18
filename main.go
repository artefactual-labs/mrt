package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"

	"github.com/containerd/go-runc"
	"github.com/opencontainers/runc/libcontainer/specconv"
	"github.com/opencontainers/runtime-spec/specs-go"

	"github.com/artefactual-labs/mrt/dist"
)

const containerID = "arbutus"

func main() {
	ctx := context.Background()

	tmpDir, err := os.MkdirTemp("", "mrt-*")
	if err != nil {
		log.Fatal(err)
	}

	runcPath, err := installRunc(tmpDir)
	if err != nil {
		log.Fatal("Failed to install runc: ", err)
	}

	bundle, err := prepareBundle(ctx, tmpDir)
	if err != nil {
		log.Fatal(err)
	}

	r := runc.Runc{Command: runcPath}

	if err = r.Delete(ctx, containerID, &runc.DeleteOpts{Force: true}); err != nil {
		log.Fatal(err)
	} else {
		log.Println("Container deleted.")
	}

	log.Println("Creating container...")
	if pid, err := r.Run(ctx, "arbutus", bundle, &runc.CreateOpts{}); err != nil {
		log.Fatal(err)
	} else {
		log.Printf("Container executed - pid %d", pid)
	}
}

func installRunc(baseDir string) (string, error) {
	return provideAsset("assets/runc.amd64", filepath.Join(baseDir, "runc"), 0o750)
}

func provideAsset(path, dest string, mode os.FileMode) (string, error) {
	src, err := dist.Assets.Open(path)
	if err != nil {
		return "", err
	}

	destFile, err := os.Create(dest)
	if err != nil {
		return "", err
	}
	defer destFile.Close()

	if err := os.Chmod(dest, mode); err != nil {
		return "", err
	}

	_, err = io.Copy(destFile, src)
	if err != nil {
		return "", err
	}

	err = destFile.Sync()
	if err != nil {
		return "", err
	}

	return dest, nil
}

func isRoot() bool {
	currentUser, err := user.Current()
	if err != nil {
		return false
	}
	return currentUser.Uid == "0"
}

func prepareSpec(dest string, rootfs string, args []string) error {
	spec := specconv.Example()
	if !isRoot() {
		specconv.ToRootless(spec)
	}

	spec.Process.Args = args
	spec.Root = &specs.Root{
		Path:     rootfs,
		Readonly: true,
	}

	if blob, err := json.Marshal(&spec); err != nil {
		return err
	} else if err := os.WriteFile(dest, blob, os.FileMode(0o660)); err != nil {
		return err
	}

	return nil
}

func prepareBundle(ctx context.Context, baseDir string) (string, error) {
	fsTar, err := provideAsset("assets/rootfs.tar", filepath.Join(baseDir, "rootfs.tar"), 0o640)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = os.Remove(fsTar)
	}()

	bundleDir := filepath.Join(baseDir, "bundle")
	configFile := filepath.Join(bundleDir, "config.json")
	rootFsDir := filepath.Join(bundleDir, "rootfs")

	if err := os.MkdirAll(rootFsDir, os.FileMode(0o750)); err != nil {
		return "", err
	}

	cmd := exec.CommandContext(ctx, "tar", "-xf", fsTar, "-C", rootFsDir)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("extract tar file: %v", err)
	}

	if err := prepareSpec(configFile, rootFsDir, []string{"sleep", "3"}); err != nil {
		return "", fmt.Errorf("write spec file: %v", err)
	}

	return bundleDir, nil
}
