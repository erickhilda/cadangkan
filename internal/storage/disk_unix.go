//go:build unix || linux || darwin

package storage

import "syscall"

// getDiskSpace returns available disk space in bytes for the given path.
func getDiskSpace(path string) (uint64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, err
	}

	// Available space = Available blocks * Block size
	availableBytes := stat.Bavail * uint64(stat.Bsize)
	return availableBytes, nil
}
