package workspace

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestPackExtract_Roundtrip(t *testing.T) {
	src := t.TempDir()
	os.WriteFile(filepath.Join(src, "hello.txt"), []byte("world"), 0644)
	sub := filepath.Join(src, "sub")
	os.MkdirAll(sub, 0755)
	os.WriteFile(filepath.Join(sub, "deep.txt"), []byte("data"), 0644)

	var buf bytes.Buffer
	if err := Pack(src, &buf); err != nil {
		t.Fatalf("pack failed: %v", err)
	}

	dst := t.TempDir()
	if err := Extract(&buf, dst); err != nil {
		t.Fatalf("extract failed: %v", err)
	}

	b, err := os.ReadFile(filepath.Join(dst, "hello.txt"))
	if err != nil {
		t.Fatalf("read hello.txt: %v", err)
	}
	if string(b) != "world" {
		t.Fatalf("hello.txt content mismatch")
	}

	b, err = os.ReadFile(filepath.Join(dst, "sub", "deep.txt"))
	if err != nil {
		t.Fatalf("read deep.txt: %v", err)
	}
	if string(b) != "data" {
		t.Fatalf("deep.txt content mismatch")
	}
}
