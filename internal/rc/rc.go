package rc

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yourusername/shctl/internal/util"
)

var (
	DefaultBackupDir = "/tmp"
)

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func RCPath() string {
	if v := getenv("BASM_RC_FILE", ""); v != "" {
		return v
	}
	home, _ := os.UserHomeDir()
	shell := getenv("SHELL", "/bin/bash")
	def := ".bashrc"
	if strings.HasSuffix(shell, "zsh") {
		def = ".zshrc"
	}
	return filepath.Join(home, def)
}

func BackupDir() string {
	if v := getenv("BASM_BACKUP_DIR", ""); v != "" {
		return v
	}
	return DefaultBackupDir
}

// Ensure rc file exists
func ensureFile() error {
	p := RCPath()
	dir := filepath.Dir(p)
	if dir == "" {
		dir = "."
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(p); errors.Is(err, os.ErrNotExist) {
		f, err := os.OpenFile(p, os.O_CREATE, 0o644)
		if err != nil {
			return err
		}
		return f.Close()
	}
	return nil
}

func AddAlias(name, command string) error {
	if err := ensureFile(); err != nil {
		return err
	}
	line := fmt.Sprintf("alias %s='%s'\n", name, command)
	return util.AppendFileAtomic(RCPath(), []byte(line))
}

func ListAliases(w io.Writer) error {
	if err := ensureFile(); err != nil {
		return err
	}
	f, err := os.Open(RCPath())
	if err != nil {
		return err
	}
	defer f.Close()
	return scanPrintPrefix(f, "alias ", w)
}

func RemoveAlias(name string) error {
	if err := ensureFile(); err != nil {
		return err
	}
	prefix := "alias " + name + "="
	return util.RemoveLinesWithPrefix(RCPath(), prefix)
}

func AddExport(varName, value string) error {
	if err := ensureFile(); err != nil {
		return err
	}
	if strings.Contains(value, " ") {
		value = fmt.Sprintf("\"%s\"", value)
	}
	line := fmt.Sprintf("export %s=%s\n", varName, value)
	return util.AppendFileAtomic(RCPath(), []byte(line))
}

func ListExports(w io.Writer) error {
	if err := ensureFile(); err != nil {
		return err
	}
	f, err := os.Open(RCPath())
	if err != nil {
		return err
	}
	defer f.Close()
	return scanPrintPrefix(f, "export ", w)
}

func RemoveExport(varName string) error {
	if err := ensureFile(); err != nil {
		return err
	}
	prefix := "export " + varName + "="
	return util.RemoveLinesWithPrefix(RCPath(), prefix)
}

func Backup(includeRC bool) error {
	dir := BackupDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if includeRC {
		src := RCPath()
		dst := filepath.Join(dir, filepath.Base(src)+".bak."+time.Now().Format("20060102_150405"))
		if err := util.CopyFile(src, dst); err != nil {
			return err
		}
	}
	return nil
}

func Restore() error {
	// Find latest rc backup in backup dir
	dir := BackupDir()
	pattern := filepath.Join(dir, filepath.Base(RCPath())+".bak.*")
	matches, _ := filepath.Glob(pattern)
	if len(matches) == 0 {
		return fmt.Errorf("no rc backup found in %s", dir)
	}
	latest := util.LatestFile(matches)
	return util.CopyFile(latest, RCPath())
}

// scanning helper
func scanPrintPrefix(r io.Reader, prefix string, w io.Writer) error {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(strings.TrimSpace(line), prefix) {
			if _, err := fmt.Fprintln(w, line); err != nil {
				return err
			}
		}
	}
	return sc.Err()
}
