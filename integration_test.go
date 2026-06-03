package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/nex1gen/sendix/internal/bundle"
	"github.com/nex1gen/sendix/internal/packer"
	"github.com/nex1gen/sendix/internal/unpacker"
)

func TestPackUnpackRoundtrip(t *testing.T) {
	repo := t.TempDir()
	os.WriteFile(filepath.Join(repo, "hello.txt"), []byte("world"), 0644)
	os.WriteFile(filepath.Join(repo, "untracked.txt"), []byte("secret"), 0644)

	if out, err := exec.Command("git", "-C", repo, "init").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	exec.Command("git", "-C", repo, "config", "user.email", "t@t.com").Run()
	exec.Command("git", "-C", repo, "config", "user.name", "T").Run()
	exec.Command("git", "-C", repo, "add", ".").Run()
	exec.Command("git", "-C", repo, "commit", "-m", "init").Run()

	archive := filepath.Join(t.TempDir(), "test.sdx")
	if err := packer.Pack(repo, archive, "testpass", false); err != nil {
		t.Fatalf("pack failed: %v", err)
	}

	dst := t.TempDir()
	if err := unpacker.Unpack(archive, dst, "testpass"); err != nil {
		t.Fatalf("unpack failed: %v", err)
	}

	b, err := os.ReadFile(filepath.Join(dst, "hello.txt"))
	if err != nil {
		t.Fatalf("read hello.txt: %v", err)
	}
	if string(b) != "world" {
		t.Fatalf("content mismatch: got %q", b)
	}

	b, err = os.ReadFile(filepath.Join(dst, "untracked.txt"))
	if err != nil {
		t.Fatalf("read untracked.txt: %v", err)
	}
	if string(b) != "secret" {
		t.Fatalf("untracked content mismatch: got %q", b)
	}

	if !bundle.IsRepo(dst) {
		t.Fatal("destination is not a git repo after unpack")
	}
}
