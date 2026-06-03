package packer

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/nex1gen/sendix/internal/bundle"
	"github.com/nex1gen/sendix/internal/crypto"
	"github.com/nex1gen/sendix/internal/workspace"
)

// Pack creates an encrypted .sdx archive from repoDir, writing to outputPath.
func Pack(repoDir, outputPath, password string, allBranches bool) error {
	if !bundle.IsRepo(repoDir) {
		return fmt.Errorf("%s is not a git repository", repoDir)
	}

	tmpDir, err := os.MkdirTemp("", "sendix-pack-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	var refs []string
	if allBranches {
		refs = []string{"--all"}
	} else {
		branch, err := bundle.CurrentBranch(repoDir)
		if err != nil {
			return fmt.Errorf("get current branch: %w", err)
		}
		refs = []string{"refs/heads/" + branch}
	}

	bundlePath := filepath.Join(tmpDir, "repo.bundle")
	if err := bundle.Create(repoDir, bundlePath, refs); err != nil {
		return fmt.Errorf("create git bundle: %w", err)
	}

	workspacePath := filepath.Join(tmpDir, "workspace.tar.gz")
	wf, err := os.Create(workspacePath)
	if err != nil {
		return fmt.Errorf("create workspace archive: %w", err)
	}

	// Avoid including any existing .sdx archives inside the repo.
	absRepo, _ := filepath.Abs(repoDir)
	absRepo = filepath.Clean(absRepo)
	var skip []string
	entries, _ := os.ReadDir(absRepo)
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".sdx") {
			skip = append(skip, filepath.Join(absRepo, e.Name()))
		}
	}

	if err := workspace.Pack(repoDir, wf, skip...); err != nil {
		wf.Close()
		return fmt.Errorf("pack workspace: %w", err)
	}
	wf.Close()

	innerPath := filepath.Join(tmpDir, "inner.tar.gz")
	innerF, err := os.Create(innerPath)
	if err != nil {
		return fmt.Errorf("create inner archive: %w", err)
	}
	gw := gzip.NewWriter(innerF)
	tw := tar.NewWriter(gw)

	if err := addFileToTar(tw, bundlePath, "repo.bundle"); err != nil {
		innerF.Close()
		return fmt.Errorf("add bundle to archive: %w", err)
	}
	if err := addFileToTar(tw, workspacePath, "workspace.tar.gz"); err != nil {
		innerF.Close()
		return fmt.Errorf("add workspace to archive: %w", err)
	}

	if err := tw.Close(); err != nil {
		innerF.Close()
		return err
	}
	if err := gw.Close(); err != nil {
		innerF.Close()
		return err
	}
	innerF.Close()

	outF, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer outF.Close()

	innerR, err := os.Open(innerPath)
	if err != nil {
		return fmt.Errorf("open inner archive: %w", err)
	}
	defer innerR.Close()

	if err := crypto.Encrypt(outF, innerR, password); err != nil {
		return fmt.Errorf("encrypt archive: %w", err)
	}
	return nil
}

func addFileToTar(tw *tar.Writer, path, name string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	header.Name = name
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(tw, f)
	return err
}
