package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCLIHelp(t *testing.T) {
	cmd := exec.Command("go", "run", "./cmd/shctl", "--help")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("help failed: %v\n%s", err, out)
	}
	if !contains(string(out), "Usage:") && !contains(string(out), "shctl") {
		t.Fatalf("unexpected help output: %s", out)
	}
}

func TestBinaryAliasAddList(t *testing.T) {
	tmp := t.TempDir()
	os.Setenv("BASM_RC_FILE", filepath.Join(tmp, "rc"))
	os.Setenv("BASM_SUDOERS_PATH", filepath.Join(tmp, "sudo"))
	os.Setenv("BASM_BACKUP_DIR", tmp)

	// build the tool
	cmd := exec.Command("go", "build", "-o", filepath.Join(tmp, "shctl"), "./cmd/shctl")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}

	bin := filepath.Join(tmp, "shctl")
	// add alias
	if out, err := exec.Command(bin, "alias", "add", "hi", "echo hi").CombinedOutput(); err != nil {
		t.Fatalf("alias add failed: %v\n%s", err, out)
	}
	// list alias
	if out, err := exec.Command(bin, "alias", "list").CombinedOutput(); err != nil {
		t.Fatalf("alias list failed: %v\n%s", err, out)
	} else if !contains(string(out), "alias hi='echo hi'") {
		t.Fatalf("unexpected alias list: %s", out)
	}
}
