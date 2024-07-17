package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/containerd/go-runc"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sevein/mrt/dist"
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

	// Generated with `runc spec --rootless`.
	config := specs.Spec{
		Version: "1.0.2",
		Process: &specs.Process{
			Terminal: false,
			User: specs.User{
				UID: 0,
				GID: 0,
			},
			Args: []string{"sleep", "3"},
			Env: []string{
				"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
				"TERM=xterm",
			},
			Cwd: "/",
			Capabilities: &specs.LinuxCapabilities{
				Bounding: []string{
					"CAP_AUDIT_WRITE",
					"CAP_KILL",
					"CAP_NET_BIND_SERVICE",
				},
				Effective: []string{
					"CAP_AUDIT_WRITE",
					"CAP_KILL",
					"CAP_NET_BIND_SERVICE",
				},
				Permitted: []string{
					"CAP_AUDIT_WRITE",
					"CAP_KILL",
					"CAP_NET_BIND_SERVICE",
				},
				Ambient: []string{
					"CAP_AUDIT_WRITE",
					"CAP_KILL",
					"CAP_NET_BIND_SERVICE",
				},
			},
			Rlimits: []specs.POSIXRlimit{
				{
					Type: "RLIMIT_NOFILE",
					Hard: 1024,
					Soft: 1024,
				},
			},
			NoNewPrivileges: true,
		},
		Root: &specs.Root{
			Path:     rootFsDir,
			Readonly: true,
		},
		Hostname: "runc",
		Mounts: []specs.Mount{
			{
				Destination: "/proc",
				Type:        "proc",
				Source:      "proc",
			},
			{
				Destination: "/dev",
				Type:        "tmpfs",
				Source:      "tmpfs",
				Options: []string{
					"nosuid",
					"strictatime",
					"mode=755",
					"size=65536k",
				},
			},
			{
				Destination: "/dev/pts",
				Type:        "devpts",
				Source:      "devpts",
				Options: []string{
					"nosuid",
					"noexec",
					"newinstance",
					"ptmxmode=0666",
					"mode=0620",
				},
			},
			{
				Destination: "/dev/shm",
				Type:        "tmpfs",
				Source:      "shm",
				Options: []string{
					"nosuid",
					"noexec",
					"nodev",
					"mode=1777",
					"size=65536k",
				},
			},
			{
				Destination: "/dev/mqueue",
				Type:        "mqueue",
				Source:      "mqueue",
				Options: []string{
					"nosuid",
					"noexec",
					"nodev",
				},
			},
			{
				Destination: "/sys",
				Type:        "none",
				Source:      "/sys",
				Options: []string{
					"rbind",
					"nosuid",
					"noexec",
					"nodev",
					"ro",
				},
			},
			{
				Destination: "/sys/fs/cgroup",
				Type:        "cgroup",
				Source:      "cgroup",
				Options: []string{
					"nosuid",
					"noexec",
					"nodev",
					"relatime",
					"ro",
				},
			},
		},
		Linux: &specs.Linux{
			Namespaces: []specs.LinuxNamespace{
				{Type: specs.PIDNamespace},
				{Type: specs.UTSNamespace},
				{Type: specs.IPCNamespace},
				{Type: specs.MountNamespace},
				{Type: specs.UserNamespace},
			},
			UIDMappings: []specs.LinuxIDMapping{
				{
					ContainerID: 0,
					HostID:      uint32(os.Geteuid()), // Map container root to the current user
					Size:        1,
				},
			},
			GIDMappings: []specs.LinuxIDMapping{
				{
					ContainerID: 0,
					HostID:      uint32(os.Getegid()), // Map container root to the current group
					Size:        1,
				},
			},
			MaskedPaths: []string{
				"/proc/acpi",
				"/proc/asound",
				"/proc/kcore",
				"/proc/keys",
				"/proc/latency_stats",
				"/proc/timer_list",
				"/proc/timer_stats",
				"/proc/sched_debug",
				"/sys/firmware",
				"/proc/scsi",
			},
			ReadonlyPaths: []string{
				"/proc/bus",
				"/proc/fs",
				"/proc/irq",
				"/proc/sys",
				"/proc/sysrq-trigger",
			},
		},
	}

	if blob, err := json.Marshal(&config); err != nil {
		return "", err
	} else if err := os.WriteFile(configFile, blob, os.FileMode(0o640)); err != nil {
		return "", err
	}

	return bundleDir, nil
}
