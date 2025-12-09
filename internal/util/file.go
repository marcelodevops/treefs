package util

import (
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// AppendFileAtomic appends bytes to file safely (open append -> write).
func AppendFileAtomic(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}

func AppendFileAtomicNoCreate(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}

// RemoveLinesWithPrefix rewrites file excluding lines that start with prefix.
func RemoveLinesWithPrefix(path, prefix string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := splitLines(string(b))
	out := []string{}
	for _, l := range lines {
		if len(l) >= len(prefix) && l[:len(prefix)] == prefix {
			continue
		}
		out = append(out, l)
	}
	return atomicWrite(path, []byte(joinLines(out)))
}

func RemoveLinesContaining(path, pattern string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := splitLines(string(b))
	out := []string{}
	for _, l := range lines {
		if contains(l, pattern) {
			continue
		}
		out = append(out, l)
	}
	return atomicWrite(path, []byte(joinLines(out)))
}

func contains(s, sub string) bool {
	return len(sub) > 0 && (len(s) >= len(sub) && (index(s, sub) >= 0))
}

// Helpers (simple to avoid imports)
func splitLines(s string) []string {
	if s == "" {
		return []string{}
	}
	return stringsSplit(s, "\n")
}

func joinLines(lines []string) string {
	return stringsJoin(lines, "\n")
}

func atomicWrite(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp := filepath.Join(dir, ".tmp_"+filepath.Base(path))
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func CopyToTemp(src string) (string, error) {
	b, err := os.ReadFile(src)
	if err != nil {
		return "", err
	}
	f, err := os.CreateTemp("", "shctl_sudoers_*")
	if err != nil {
		return "", err
	}
	if _, err := f.Write(b); err != nil {
		f.Close()
		return "", err
	}
	f.Close()
	if fi, err := os.Stat(src); err == nil {
		_ = os.Chmod(f.Name(), fi.Mode())
	}
	return f.Name(), nil
}

func LatestFile(files []string) string {
	sort.Slice(files, func(i, j int) bool {
		ti := modTime(files[i])
		tj := modTime(files[j])
		return ti.After(tj)
	})
	return files[0]
}

func modTime(path string) time.Time {
	if fi, err := os.Stat(path); err == nil {
		return fi.ModTime()
	}
	return time.Time{}
}
