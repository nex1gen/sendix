package unpacker

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nex1gen/sendix/internal/bundle"
	"github.com/nex1gen/sendix/internal/crypto"
	"github.com/nex1gen/sendix/internal/workspace"
)

// Unpack decrypts and extracts an .sdx archive into destDir.
func Unpack(archivePath, destDir, password string) error {
	tmpDir, err := os.MkdirTemp("", "sendix-unpack-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	innerPath := filepath.Join(tmpDir, "inner.tar.gz")
	innerF, err := os.Create(innerPath)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	arcF, err := os.Open(archivePath)
	if err != nil {
		innerF.Close()
		return fmt.Errorf("open archive: %w", err)
	}
	if err := crypto.Decrypt(innerF, arcF, password); err != nil {
		arcF.Close()
		innerF.Close()
		return fmt.Errorf("decrypt archive: %w", err)
	}
	arcF.Close()
	innerF.Close()

	innerR, err := os.Open(innerPath)
	if err != nil {
		return fmt.Errorf("open decrypted archive: %w", err)
	}
	defer innerR.Close()

	gr, err := gzip.NewReader(innerR)
	if err != nil {
		return fmt.Errorf("gzip reader: %w", err)
	}
	defer gr.Close()
	tr := tar.NewReader(gr)

	var bundlePath, wsPath string
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read inner tar: %w", err)
		}
		switch header.Name {
		case "repo.bundle":
			bundlePath = filepath.Join(tmpDir, "repo.bundle")
			if err := extractFile(tr, bundlePath); err != nil {
				return fmt.Errorf("extract bundle: %w", err)
			}
		case "workspace.tar.gz":
			wsPath = filepath.Join(tmpDir, "workspace.tar.gz")
			if err := extractFile(tr, wsPath); err != nil {
				return fmt.Errorf("extract workspace: %w", err)
			}
		}
	}

	if bundlePath == "" {
		return fmt.Errorf("archive missing repo.bundle")
	}
	if wsPath == "" {
		return fmt.Errorf("archive missing workspace.tar.gz")
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("create destination: %w", err)
	}

	hasGit := bundle.IsRepo(destDir)
	if !hasGit {
		if err := initAndUnbundle(destDir, bundlePath); err != nil {
			return fmt.Errorf("init and unbundle: %w", err)
		}
	} else {
		if err := bundle.Unbundle(destDir, bundlePath); err != nil {
			return fmt.Errorf("unbundle: %w", err)
		}
	}

	wsF, err := os.Open(wsPath)
	if err != nil {
		return fmt.Errorf("open workspace archive: %w", err)
	}
	defer wsF.Close()
	if err := workspace.Extract(wsF, destDir); err != nil {
		return fmt.Errorf("extract workspace: %w", err)
	}

	return nil
}

func initAndUnbundle(destDir, bundlePath string) error {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}
	// Initialize empty repo with a dummy branch so fetch can overwrite real ones
	cmd := exec.Command("git", "init")
	cmd.Dir = destDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git init failed: %w", err)
	}
	cmd = exec.Command("git", "symbolic-ref", "HEAD", "refs/heads/dummy")
	cmd.Dir = destDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git symbolic-ref failed: %w", err)
	}
	// Fetch imports objects and refs from bundle
	cmd = exec.Command("git", "fetch", bundlePath, "refs/*:refs/*")
	cmd.Dir = destDir
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git fetch from bundle failed: %w", err)
	}
	// Get the first ref from bundle and checkout
	heads, err := bundle.ListHeads(bundlePath)
	if err != nil {
		return err
	}
	if len(heads) == 0 {
		return fmt.Errorf("bundle contains no refs")
	}
	ref := heads[0]
	ref = strings.TrimPrefix(ref, "refs/heads/")
	if err := bundle.Checkout(destDir, ref); err != nil {
		return err
	}
	return nil
}

func extractFile(r io.Reader, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}
