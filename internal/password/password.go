package password

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// Read prompts for a password with hidden input.
// If SENDIX_PASSWORD env var is set, returns it directly (useful for CI/testing).
func Read(prompt string) (string, error) {
	if env := os.Getenv("SENDIX_PASSWORD"); env != "" {
		return env, nil
	}
	fmt.Fprint(os.Stderr, prompt)
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	return string(b), nil
}

// ReadConfirmed prompts twice and ensures they match.
// If SENDIX_PASSWORD env var is set, returns it directly without confirmation.
func ReadConfirmed() (string, error) {
	if env := os.Getenv("SENDIX_PASSWORD"); env != "" {
		return env, nil
	}
	p1, err := Read("Enter password: ")
	if err != nil {
		return "", err
	}
	p2, err := Read("Confirm password: ")
	if err != nil {
		return "", err
	}
	if p1 != p2 {
		return "", fmt.Errorf("passwords do not match")
	}
	return p1, nil
}
