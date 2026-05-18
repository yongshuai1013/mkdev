//go:build !windows

package safeexec

import (
	"fmt"
	"os"
	"syscall"
)

func verifyOwner(resolved string, info os.FileInfo) error {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return nil
	}
	uid := stat.Uid
	if uid != 0 && int(uid) != os.Getuid() {
		return fmt.Errorf("safeexec: %s owned by uid %d (not root or current user); refusing to invoke under sudo", resolved, uid)
	}
	return nil
}
