// Package fsys is the filesystem every feature writes through, so --dry-run and the
// unit tests share one code path with a real run.
package fsys

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FS is the subset of filesystem operations forgego needs. Paths are slash- or
// OS-separated; implementations normalise them.
type FS interface {
	Exists(path string) bool
	ReadFile(path string) (string, error)
	WriteFile(path, content string) error
	Remove(path string) error
}

// OS writes to the real filesystem, creating parent directories as needed.
type OS struct{}

func (OS) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (OS) ReadFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	return string(b), err
}

func (OS) WriteFile(path, content string) error {
	// Scaffolded files are committed to the project's repository, so they're world-readable.
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil { //nolint:gosec // see above
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644) //nolint:gosec // committed to the repository
}

func (OS) Remove(path string) error {
	return os.Remove(path)
}

// Memory is an in-memory filesystem for fast, deterministic tests.
type Memory struct {
	files map[string]string
}

// NewMemory returns a Memory filesystem seeded with files, keyed by path.
func NewMemory(files map[string]string) *Memory {
	m := &Memory{files: map[string]string{}}
	for path, content := range files {
		m.files[key(path)] = content
	}
	return m
}

func key(path string) string {
	return filepath.ToSlash(filepath.Clean(path))
}

func (m *Memory) Exists(path string) bool {
	k := key(path)
	if _, ok := m.files[k]; ok {
		return true
	}
	prefix := k + "/"
	for name := range m.files {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

func (m *Memory) ReadFile(path string) (string, error) {
	content, ok := m.files[key(path)]
	if !ok {
		return "", &fs.PathError{Op: "open", Path: path, Err: fs.ErrNotExist}
	}
	return content, nil
}

func (m *Memory) WriteFile(path, content string) error {
	m.files[key(path)] = content
	return nil
}

func (m *Memory) Remove(path string) error {
	k := key(path)
	if _, ok := m.files[k]; !ok {
		return &fs.PathError{Op: "remove", Path: path, Err: fs.ErrNotExist}
	}
	delete(m.files, k)
	return nil
}

// Paths lists every file, sorted, for assertions.
func (m *Memory) Paths() []string {
	paths := make([]string, 0, len(m.files))
	for path := range m.files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

// IsNotExist reports whether err means the file doesn't exist.
func IsNotExist(err error) bool {
	return errors.Is(err, fs.ErrNotExist)
}
