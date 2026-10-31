// Package credentials provides protected local credential stores. The
// interface keeps the CLI independent from macOS Keychain libraries; a
// Keychain implementation can be supplied by the platform package later.
package credentials

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var ErrNotFound = errors.New("credential not found")
var ErrKeychainUnavailable = errors.New("macOS keychain is unavailable in this build")

type Store interface {
	Get(name string) (string, error)
	Set(name, value string) error
	Delete(name string) error
}

// FileStore is the documented headless fallback. Files are kept mode 0600
// and writes are atomic enough for a single-user launchd process.
type FileStore struct{ Dir string }

func (s FileStore) path(name string) string {
	clean := filepath.Base(name)
	return filepath.Join(s.Dir, clean)
}

func (s FileStore) Get(name string) (string, error) {
	b, err := os.ReadFile(s.path(name))
	if os.IsNotExist(err) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("read credential: %w", err)
	}
	v := strings.TrimSpace(string(b))
	if v == "" {
		return "", ErrNotFound
	}
	return v, nil
}

func (s FileStore) Set(name, value string) error {
	if strings.TrimSpace(value) == "" {
		return ErrNotFound
	}
	if err := os.MkdirAll(s.Dir, 0o700); err != nil {
		return fmt.Errorf("create credential directory: %w", err)
	}
	tmp, err := os.CreateTemp(s.Dir, ".credential-*")
	if err != nil {
		return fmt.Errorf("create credential temporary file: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.WriteString(strings.TrimSpace(value) + "\n"); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, s.path(name)); err != nil {
		return fmt.Errorf("replace credential: %w", err)
	}
	return nil
}

func (s FileStore) Delete(name string) error {
	err := os.Remove(s.path(name))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// KeychainStore is an explicit boundary for the future macOS implementation.
// Returning an error is safer than silently falling back to an insecure store.
type KeychainStore struct{}

func (KeychainStore) Get(string) (string, error) { return "", ErrKeychainUnavailable }
func (KeychainStore) Set(string, string) error   { return ErrKeychainUnavailable }
func (KeychainStore) Delete(string) error        { return ErrKeychainUnavailable }
