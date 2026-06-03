package crypto

import (
	"bytes"
	"strings"
	"testing"
)

func TestEncryptDecrypt_Roundtrip(t *testing.T) {
	plain := []byte("hello world, this is a secret message")
	password := "super-secret-password-123"

	var encrypted bytes.Buffer
	if err := Encrypt(&encrypted, bytes.NewReader(plain), password); err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	var decrypted bytes.Buffer
	if err := Decrypt(&decrypted, &encrypted, password); err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}

	if !bytes.Equal(decrypted.Bytes(), plain) {
		t.Fatalf("decrypted mismatch: got %q, want %q", decrypted.Bytes(), plain)
	}
}

func TestDecrypt_WrongPassword(t *testing.T) {
	plain := []byte("secret")
	var encrypted bytes.Buffer
	if err := Encrypt(&encrypted, bytes.NewReader(plain), "right"); err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	var decrypted bytes.Buffer
	err := Decrypt(&decrypted, &encrypted, "wrong")
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
	if !strings.Contains(err.Error(), "invalid password") {
		t.Fatalf("unexpected error: %v", err)
	}
}
