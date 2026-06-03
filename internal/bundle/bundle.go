package bundle

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Create runs `git bundle create` in the given repo directory.
func Create(repoDir, outputPath string, refs []string) error {
	args := []string{"bundle", "create", outputPath}
	args = append(args, refs...)
	cmd := exec.Command("git", args...)
	cmd.Dir = repoDir
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git bundle create failed: %w", err)
	}
	return nil
}

// CurrentBranch returns the current branch name, or "HEAD" if detached.
func CurrentBranch(repoDir string) (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = repoDir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git branch --show-current failed: %w", err)
	}
	branch := strings.TrimSpace(string(out))
	if branch == "" {
		return "HEAD", nil
	}
	return branch, nil
}

// Unbundle runs `git bundle unbundle` in the given repo directory.
func Unbundle(repoDir, bundlePath string) error {
	cmd := exec.Command("git", "bundle", "unbundle", bundlePath)
	cmd.Dir = repoDir
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git bundle unbundle failed: %w", err)
	}
	return nil
}

// ListHeads returns the refs contained in a bundle.
func ListHeads(bundlePath string) ([]string, error) {
	cmd := exec.Command("git", "bundle", "list-heads", bundlePath)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git bundle list-heads failed: %w", err)
	}
	var heads []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			heads = append(heads, parts[1])
		}
	}
	return heads, nil
}

// Checkout runs `git checkout` in the given repo directory.
func Checkout(repoDir, ref string) error {
	cmd := exec.Command("git", "checkout", ref)
	cmd.Dir = repoDir
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git checkout %s failed: %w", ref, err)
	}
	return nil
}

// IsRepo returns true if dir is a git repository.
func IsRepo(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil
}
