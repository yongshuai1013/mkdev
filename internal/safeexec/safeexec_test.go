package safeexec_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/venkatkrishna07/mkdev/internal/safeexec"
)

func TestVerifyBinPath(t *testing.T) {
	dir := t.TempDir()

	okPath := filepath.Join(dir, "ok")
	require.NoError(t, os.WriteFile(okPath, []byte("#!/bin/sh\nexit 0\n"), 0o755))

	groupWritablePath := filepath.Join(dir, "groupw")
	require.NoError(t, os.WriteFile(groupWritablePath, []byte("x"), 0o775))
	require.NoError(t, os.Chmod(groupWritablePath, 0o775))

	worldWritablePath := filepath.Join(dir, "worldw")
	require.NoError(t, os.WriteFile(worldWritablePath, []byte("x"), 0o757))
	require.NoError(t, os.Chmod(worldWritablePath, 0o757))

	symlinkPath := filepath.Join(dir, "linked")
	require.NoError(t, os.Symlink(okPath, symlinkPath))

	notRegularPath := filepath.Join(dir, "asdir")
	require.NoError(t, os.Mkdir(notRegularPath, 0o755))

	// Unix-only perm-bit assertions; Windows uses ACLs.
	unixOnly := runtime.GOOS != "windows"

	cases := []struct {
		name    string
		path    string
		wantErr bool
		skip    bool
	}{
		{name: "regular owned by current uid", path: okPath, wantErr: false},
		{name: "group writable rejected", path: groupWritablePath, wantErr: true, skip: !unixOnly},
		{name: "world writable rejected", path: worldWritablePath, wantErr: true, skip: !unixOnly},
		{name: "symlink follows to ok target", path: symlinkPath, wantErr: false},
		{name: "directory rejected", path: notRegularPath, wantErr: true},
		{name: "missing file rejected", path: filepath.Join(dir, "nope"), wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.skip {
				t.Skip("Unix-only perm bit check")
			}
			err := safeexec.VerifyBinPath(tc.path)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestVerifyBinPathForeignUid is platform-conditional. We can't chown to
// uid=0 from a test without root, but we can validate the rejection path
// only when running as non-root by symlinking to /bin/ls (root-owned and
// safely permissioned) — that should always succeed.
func TestVerifyBinPathRootOwnedSystemBinary(t *testing.T) {
	if _, err := os.Stat("/bin/ls"); err != nil {
		t.Skip("/bin/ls not present")
	}
	require.NoError(t, safeexec.VerifyBinPath("/bin/ls"))
}
