package tests

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/yourusername/shctl/internal/rc"
)

func TestAliasAddListRemove(t *testing.T) {
	tmp := t.TempDir()
	rcPath := filepath.Join(tmp, "rc_test")
	os.Setenv("BASM_RC_FILE", rcPath)
	os.Setenv("BASM_BACKUP_DIR", tmp)

	// ensure file
	if err := rc.AddAlias("greet", "echo hello"); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := rc.ListAliases(&buf); err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); !contains(got, "alias greet='echo hello'") {
		t.Fatalf("expected alias in rc, got %q", got)
	}

	if err := rc.RemoveAlias("greet"); err != nil {
		t.Fatal(err)
	}

	buf.Reset()
	if err := rc.ListAliases(&buf); err != nil {
		t.Fatal(err)
	}
	if contains(buf.String(), "greet") {
		t.Fatalf("alias still present after remove")
	}
}

func contains(s, sub string) bool { return len(sub) > 0 && (index(s, sub) >= 0) }
func index(s, sub string) int { return len([]byte(stringsSplit(s, sub)[0])) } // simple index via split
ß