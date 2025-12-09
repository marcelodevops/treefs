package sudoers

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/yourusername/shctl/internal/util"
)

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func SudoersPath() string {
	if v := getenv("BASM_SUDOERS_PATH", ""); v != "" {
		return v
	}
	return "/etc/sudoers"
}

func BackupDir() string {
	if v := getenv("BASM_BACKUP_DIR", ""); v != "" {
		return v
	}
	return "/tmp"
}

func List(w io.Writer) error {
	f, err := os.Open(SudoersPath())
	if err != nil {
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		s := strings.TrimSpace(line)
		if s == "" || strings.HasPrefix(s, "#") {
			continue
		}
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}
	return sc.Err()
}

func Add(entry string) error {
	orig := SudoersPath()
	tmp, err := util.CopyToTemp(orig)
	if err != nil {
		return err
	}
	defer os.Remove(tmp)

	if err := util.AppendFileAtomic(tmp, []byte("\n"+entry+"\n")); err != nil {
		return err
	}
	if err := visudoValidate(tmp); err != nil {
		return fmt.Errorf("visudo validation failed: %w", err)
	}
	return copyBack(tmp, orig)
}

func Remove(pattern string) error {
	orig := SudoersPath()
	tmp, err := util.CopyToTemp(orig)
	if err != nil {
		return err
	}
	defer os.Remove(tmp)

	if err := util.RemoveLinesContaining(tmp, pattern); err != nil {
		return err
	}
	if err := visudoValidate(tmp); err != nil {
		return fmt.Errorf("visudo validation failed after removal: %w", err)
	}
	return copyBack(tmp, orig)
}

func Backup() error {
	dir := BackupDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	src := SudoersPath()
	dst := filepath.Join(dir, filepath.Base(src)+".bak."+time.Now().Format("20060102_150405"))
	return util.CopyFile(src, dst)
}

func Restore() error {
	dir := BackupDir()
	pattern := filepath.Join(dir, filepath.Base(SudoersPath())+".bak.*")
	matches, _ := filepath.Glob(pattern)
	if len(matches) == 0 {
		return fmt.Errorf("no sudoers backup found in %s", dir)
	}
	latest := util.LatestFile(matches)
	// Validate before applying
	tmp, err := util.CopyToTemp(latest)
	if err != nil {
		return err
	}
	defer os.Remove(tmp)
	if err := visudoValidate(tmp); err != nil {
		return fmt.Errorf("backup sudoers failed validation: %w", err)
	}
	return copyBack(tmp, SudoersPath())
}

func visudoValidate(path string) error {
	cmd := exec.Command("visudo", "-c", "-f", path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("visudo: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func copyBack(tmp, dest string) error {
	if dest == "/etc/sudoers" {
		// use sudo cp so file ownership/permissions preserved
		c := exec.Command("sudo", "cp", tmp, dest)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	}
	// normal copy
	return util.CopyFile(tmp, dest)
}
