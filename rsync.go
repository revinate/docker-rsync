package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var lastSyncError = ""

func Sync(via string, c SSHCredentials, src, dst string, verbose bool) {
	args := []string{
		// "--verbose",
		// "--stats",
		"--perms",
		"--chmod=a+rwx",
		"--recursive",
		"--links",
		"--times",
		"--inplace",
		"--itemize-changes",
		"--delete",
		"--force",
		"--executability",
		"--compress",
	}

	ripath := getRsyncIgnorePath(src)
	if ripath != "" {
		args = append(args, `--exclude-from='`+ripath+`'`)
	}

	if strings.HasPrefix(via, "rsync://") {
		args = append(args, filepath.Join(src)+"/.")
		args = append(args, via)
	} else {
		args = append(args, fmt.Sprintf(`-e 'ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o LogLevel=quiet -i "%s" -p %v'`, c.SSHKeyPath, c.SSHGuestPort))
		args = append(args, "--rsync-path='sudo rsync'")
		args = append(args, src, fmt.Sprintf("%s@%s:%s", c.SSHUser, c.IPAddress, dst))
	}

	cmd := Exec("rsync", args...)

	if verbose {
		fmt.Printf("rsync %v\n", strings.Join(args, " "))
		cmd.Stdout = os.Stdout
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// don't show duplicate errors
		if lastSyncError != err.Error() {
			fmt.Printf("error: %v\n", err)
		}
		lastSyncError = err.Error()
	}
}

func getRsyncIgnorePath(src string) string {
	if _, err := os.Stat(src + "docker/.rsyncignore"); err == nil {
		abs, err := filepath.Abs(src + "docker/.rsyncignore")
		if err == nil {
			return abs
		}
	}
	return ""
}
