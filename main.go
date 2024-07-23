package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"

	"github.com/containerd/go-runc"
	"github.com/go-logr/logr"
	"github.com/opencontainers/runc/libcontainer/specconv"
	"github.com/opencontainers/runtime-spec/specs-go"
	"go.artefactual.dev/tools/log"

	"github.com/artefactual-labs/mrt/dist"
)

const (
	appName     = "mrt"
	containerID = "arbutus"
)

func main() {
	ctx := context.Background()

	logger := log.New(os.Stderr,
		log.WithName(appName),
		log.WithDebug(true),
		log.WithLevel(10),
	)
	defer log.Sync(logger)

	cacheDir, err := cacheDir()
	if err != nil {
		logger.Error(err, "Failed to configure user cache directory.")
		os.Exit(1)
	}

	runcPath, err := installRunc(cacheDir)
	if err != nil {
		logger.Error(err, "Failed to install runc.")
		os.Exit(1)
	}

	args := []string{"python", "-V"}
	bundle, err := prepareBundle(ctx, logger, cacheDir, args)
	if err != nil {
		logger.Error(err, "Failed to prepare OCI bundle.")
		os.Exit(1)
	}

	r := runc.Runc{
		Command: runcPath,
	}

	ver, err := r.Version(ctx)
	if err != nil {
		logger.Error(err, "Failed to read runc version.")
		os.Exit(1)
	}
	logger.Info("Using runc.", "version", ver.Runc, "path", runcPath)

	if err = r.Delete(ctx, containerID, &runc.DeleteOpts{Force: true}); err != nil {
		logger.Error(err, "Failed to delete existing container.", "id", containerID)
		os.Exit(1)
	} else {
		logger.Info("Container deleted.", "id", containerID)
	}

	io, err := runc.NewSTDIO()
	if err != nil {
		logger.Error(err, "Failed to configure the standard streams.", "id", containerID)
		os.Exit(1)
	}

	logger.Info("Creating container", "id", containerID)
	if pid, err := r.Run(ctx, containerID, bundle, &runc.CreateOpts{
		IO: io,
	}); err != nil {
		logger.Error(err, "Failed to run container.", "id", containerID)
		os.Exit(1)
	} else {
		logger.Info("Container executed!", "pid", pid)
	}
}

func cacheDir() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(cacheDir, appName)
	if err := os.MkdirAll(path, os.FileMode(0o700)); err != nil {
		return "", err
	}

	return path, nil
}

func installRunc(cacheDir string) (string, error) {
	dest := filepath.Join(cacheDir, "runc")
	if err := dist.WriteRunc(dest); err != nil {
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

func prepareBundle(ctx context.Context, logger logr.Logger, cacheDir string, args []string) (string, error) {
	var (
		bundleDir  = filepath.Join(cacheDir, "bundle")       // » ~/.cache/mrt/bundle
		configFile = filepath.Join(bundleDir, "config.json") // » ~/.cache/mrt/bundle/config.json
		rootFsDir  = filepath.Join(bundleDir, "rootfs")      // » ~/.cache/mrt/bundle/rootfs
	)

	if err := os.MkdirAll(bundleDir, 0o750); err != nil {
		return "", fmt.Errorf("create bundle dir: %v", err)
	}

	if err := prepareSpec(configFile, rootFsDir, args); err != nil {
		return "", fmt.Errorf("write spec file: %v", err)
	}

	if err := prepareRootFS(ctx, logger, cacheDir, rootFsDir); err != nil {
		return "", fmt.Errorf("build rootfs: %v", err)
	}

	return bundleDir, nil
}

func cachedRootFS(cacheDir, dest string) bool {
	info, err := os.Stat(dest)
	if err != nil || !info.IsDir() {
		return false
	}

	sum, err := os.ReadFile(filepath.Join(cacheDir, "rootfs.tar.zst.md5"))
	if err != nil {
		return false
	}

	return dist.CheckRootFSSum(sum)
}

// prepareRootFS unpacks the rootfs.
func prepareRootFS(ctx context.Context, logger logr.Logger, cacheDir, dest string) error {
	if cachedRootFS(cacheDir, dest) {
		logger.Info("Using cached rootfs.")
		return nil
	}

	tarFile := filepath.Join(cacheDir, "rootfs.tar.zst")
	err := dist.WriteRootFS(tarFile)
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(tarFile)
	}()
	if err := os.MkdirAll(dest, os.FileMode(0o750)); err != nil {
		return err
	}

	logger.Info("Extracting rootfs.")
	cmd := exec.CommandContext(ctx, "tar", "-I", "zstd", "-xf", tarFile, "-C", dest)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("extract tar file: %v", err)
	}

	sumFile := filepath.Join(cacheDir, "rootfs.tar.zst.md5")
	if err := dist.WriteRootFSSum(sumFile); err != nil {
		return err
	}

	return nil
}
