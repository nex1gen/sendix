package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

const (
	saltLen  = 16
	nonceLen = 12
	iter     = 100000
	keyLen   = 32
)

// Encrypt writes ciphertext to w. Format: salt(16) || nonce(12) || AES-GCM(ciphertext).
func Encrypt(w io.Writer, r io.Reader, password string) error {
	salt := make([]byte, saltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return fmt.Errorf("generate salt: %w", err)
	}
	nonce := make([]byte, nonceLen)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("generate nonce: %w", err)
	}

	key := pbkdf2.Key([]byte(password), salt, iter, keyLen, sha256.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("create cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("create gcm: %w", err)
	}

	if _, err := w.Write(salt); err != nil {
		return err
	}
	if _, err := w.Write(nonce); err != nil {
		return err
	}

	plain, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("read plaintext: %w", err)
	}
	cipherText := aead.Seal(nil, nonce, plain, nil)
	if _, err := w.Write(cipherText); err != nil {
		return fmt.Errorf("write ciphertext: %w", err)
	}
	return nil
}

// Decrypt reads ciphertext from r and writes plaintext to w.
func Decrypt(w io.Writer, r io.Reader, password string) error {
	salt := make([]byte, saltLen)
	if _, err := io.ReadFull(r, salt); err != nil {
		return fmt.Errorf("read salt: %w", err)
	}
	nonce := make([]byte, nonceLen)
	if _, err := io.ReadFull(r, nonce); err != nil {
		return fmt.Errorf("read nonce: %w", err)
	}

	key := pbkdf2.Key([]byte(password), salt, iter, keyLen, sha256.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("create cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("create gcm: %w", err)
	}

	cipherText, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("read ciphertext: %w", err)
	}
	plain, err := aead.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return fmt.Errorf("invalid password or corrupted archive")
	}
	if _, err := w.Write(plain); err != nil {
		return fmt.Errorf("write plaintext: %w", err)
	}
	return nil
}
