//go:build windows

package safeexec

import "os"

// Windows ACL ownership doesn't map cleanly to uid; trust the perm-bit and
// regular-file checks done in safeexec.go.
func verifyOwner(_ string, _ os.FileInfo) error { return nil }
