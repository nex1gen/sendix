package workspace

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// excluded returns true if the path should not be included in the workspace archive.
func excluded(relPath string) bool {
	if strings.HasPrefix(relPath, ".git/objects/") {
		return true
	}
	if strings.HasPrefix(relPath, ".git/refs/") {
		return true
	}
	if strings.HasPrefix(relPath, ".git/logs/") {
		return true
	}
	if strings.Contains(relPath, ".git/HEAD.lock") || strings.Contains(relPath, ".git/index.lock") {
		return true
	}
	return false
}

// Pack creates a gzipped tar archive of srcDir, writing to w.
// Files listed in skipPaths (absolute) are omitted.
func Pack(srcDir string, w io.Writer, skipPaths ...string) error {
	gw := gzip.NewWriter(w)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	skip := make(map[string]struct{}, len(skipPaths))
	for _, p := range skipPaths {
		skip[p] = struct{}{}
	}

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		if excluded(rel) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		absPath, _ := filepath.Abs(path)
		absPath = filepath.Clean(absPath)
		if _, ok := skip[absPath]; ok {
			return nil
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("tar header for %s: %w", rel, err)
		}
		header.Name = filepath.ToSlash(rel)
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("write header for %s: %w", rel, err)
		}
		if !info.IsDir() {
			f, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("open %s: %w", rel, err)
			}
			if _, err := io.Copy(tw, f); err != nil {
				f.Close()
				return fmt.Errorf("copy %s: %w", rel, err)
			}
			f.Close()
		}
		return nil
	})
}

// Extract unpacks a gzipped tar archive from r into destDir.
func Extract(r io.Reader, destDir string) error {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("gzip reader: %w", err)
	}
	defer gr.Close()
	tr := tar.NewReader(gr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar next: %w", err)
		}
		path := filepath.Join(destDir, filepath.FromSlash(header.Name))
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("mkdir %s: %w", path, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return fmt.Errorf("mkdir parent %s: %w", path, err)
			}
			f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("create file %s: %w", path, err)
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return fmt.Errorf("copy %s: %w", path, err)
			}
			f.Close()
		default:
			// Skip unsupported types
		}
	}
	return nil
}
