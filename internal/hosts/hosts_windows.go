//go:build windows

package hosts

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/venkatkrishna07/mkdev/internal/safeexec"
)

const HostsPath = `C:\Windows\System32\drivers\etc\hosts`

type Editor struct {
	binPath string
	useGUI  bool
}

func NewEditor(mkdevBin string) *Editor {
	return &Editor{binPath: mkdevBin}
}

func NewGUIEditor(mkdevBin string) *Editor {
	return &Editor{binPath: mkdevBin, useGUI: true}
}

func (e *Editor) Read() (string, error) {
	b, err := os.ReadFile(HostsPath)
	if err != nil {
		return "", fmt.Errorf("hosts: read: %w", err)
	}
	return string(b), nil
}

func (e *Editor) Add(host string) error {
	if !ValidHostname(host) {
		return fmt.Errorf("hosts: invalid hostname %q", host)
	}
	if err := verifyBinPath(e.binPath); err != nil {
		return err
	}
	return e.runElevated("add", host)
}

func (e *Editor) Remove(host string) error {
	if !ValidHostname(host) {
		return fmt.Errorf("hosts: invalid hostname %q", host)
	}
	if err := verifyBinPath(e.binPath); err != nil {
		return err
	}
	return e.runElevated("remove", host)
}

// runElevated invokes the helper via PowerShell Start-Process -Verb RunAs,
// which triggers the UAC consent prompt. Wait blocks until the child exits.
// We reject any quote/backtick in inputs so the PowerShell literal stays
// unambiguous; hostnames are already regex-validated by ValidHostname and
// binPath by safeexec.VerifyBinPath.
func (e *Editor) runElevated(op, host string) error {
	if strings.ContainsAny(e.binPath, "\"'`") || strings.ContainsAny(host, "\"'`") {
		return fmt.Errorf("hosts: refusing to elevate with quoted path/host")
	}
	args := fmt.Sprintf("hosts-helper %s %s", op, host)
	ps := fmt.Sprintf(
		`Start-Process -FilePath '%s' -ArgumentList '%s' -Verb RunAs -Wait`,
		e.binPath, args,
	)
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", ps)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("powershell elevate: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func verifyBinPath(bin string) error {
	return safeexec.VerifyBinPath(bin)
}
