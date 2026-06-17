package objectstore

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Dir stores objects on the local filesystem rooted at the given directory.
//
// Object names are slash-separated keys. Directories are implementation detail;
// callers should treat this as object storage, not as a general filesystem.
type Dir string

var _ ObjectStore = Dir("")

func (d Dir) Get(name string) (io.ReadCloser, error) {
	full, err := d.abs(name, false)
	if err != nil {
		return nil, err
	}
	return os.Open(full)
}

func (d Dir) Put(name string, r io.Reader) error {
	full, err := d.abs(name, false)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	f, err := os.Create(full)
	if err != nil {
		return err
	}
	ok := false
	defer func() {
		if !ok {
			_ = f.Close()
		}
	}()
	if _, err := io.Copy(f, r); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	ok = true
	return nil
}

func (d Dir) Delete(name string) error {
	full, err := d.abs(name, false)
	if err != nil {
		return err
	}
	err = os.Remove(full)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (d Dir) DeletePrefix(prefix string) error {
	prefix, err := cleanName(prefix, true)
	if err != nil {
		return err
	}
	if prefix == "" {
		return nil
	}
	full := d.join(prefix)
	err = os.RemoveAll(full)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (d Dir) List(prefix string) ([]ObjectInfo, error) {
	prefix, err := cleanName(prefix, true)
	if err != nil {
		return nil, err
	}
	root := d.join(prefix)
	var out []ObjectInfo
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(d.root(), path)
		if err != nil {
			return err
		}
		out = append(out, ObjectInfo{Name: filepath.ToSlash(rel), Size: info.Size()})
		return nil
	})
	if os.IsNotExist(err) {
		return nil, nil
	}
	return out, err
}

func (d Dir) LocalDir() (string, bool) {
	return d.root(), true
}

func (d Dir) abs(name string, allowEmpty bool) (string, error) {
	name, err := cleanName(name, allowEmpty)
	if err != nil {
		return "", err
	}
	return d.join(name), nil
}

func (d Dir) join(name string) string {
	if name == "" {
		return d.root()
	}
	return filepath.Join(d.root(), filepath.FromSlash(name))
}

func (d Dir) root() string {
	if d == "" {
		return "."
	}
	return string(d)
}

func cleanName(name string, allowEmpty bool) (string, error) {
	if name == "" {
		if allowEmpty {
			return "", nil
		}
		return "", fmt.Errorf("objectstore: object name is empty")
	}
	if strings.HasPrefix(name, "/") || filepath.IsAbs(filepath.FromSlash(name)) {
		return "", fmt.Errorf("objectstore: invalid absolute object name %q", name)
	}

	parts := strings.Split(filepath.ToSlash(name), "/")
	out := parts[:0]
	for _, part := range parts {
		switch part {
		case "", ".":
			continue
		case "..":
			return "", fmt.Errorf("objectstore: invalid object name %q", name)
		default:
			out = append(out, part)
		}
	}
	if len(out) == 0 {
		if allowEmpty {
			return "", nil
		}
		return "", fmt.Errorf("objectstore: object name is empty")
	}
	name = strings.Join(out, "/")
	if name == "." || name == ".." || strings.HasPrefix(name, "../") {
		return "", fmt.Errorf("objectstore: invalid object name %q", name)
	}
	if filepath.IsAbs(filepath.FromSlash(name)) {
		return "", fmt.Errorf("objectstore: invalid absolute object name %q", name)
	}
	return name, nil
}
