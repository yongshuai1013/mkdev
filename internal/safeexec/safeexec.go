// Package safeexec validates binaries that are about to be invoked under
// elevated privileges (sudo / osascript / pkexec / UAC). It rejects paths
// in user-writable locations or owned by foreign uids, closing the obvious
// privilege-escalation hole where a malicious binary in $PATH inherits
// root via mkdev's helper invocations.
package safeexec

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// VerifyBinPath rejects bin if it is not a regular file, is group/world
// writable, or is owned by a foreign uid — closing the obvious
// privilege-escalation hole when bin is about to be invoked via sudo.
func VerifyBinPath(bin string) error {
	resolved, err := filepath.EvalSymlinks(bin)
	if err != nil {
		return fmt.Errorf("safeexec: resolve %s: %w", bin, err)
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return fmt.Errorf("safeexec: stat %s: %w", resolved, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("safeexec: %s is not a regular file", resolved)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0o022 != 0 {
		return fmt.Errorf("safeexec: %s is group/world writable; refusing to invoke under sudo", resolved)
	}
	return verifyOwner(resolved, info)
}
