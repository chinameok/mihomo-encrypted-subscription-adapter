// Package runtime owns the adapter's local state and atomic provider updates.
package runtime

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Metadata struct {
	NodeCount int    `json:"node_count"`
	Digest    string `json:"digest"`
}

type Status struct {
	Enabled      bool      `json:"enabled"`
	Provider     string    `json:"provider"`
	NodeCount    int       `json:"node_count,omitempty"`
	Digest       string    `json:"digest,omitempty"`
	LastSuccess  time.Time `json:"last_success,omitempty"`
	LastFailure  time.Time `json:"last_failure,omitempty"`
	FailureClass string    `json:"failure_class,omitempty"`
}

type Manager struct{ Dir string }

func (m Manager) providerPath() string { return filepath.Join(m.Dir, "provider.yaml") }
func (m Manager) previousPath() string { return filepath.Join(m.Dir, "provider.previous.yaml") }
func (m Manager) metadataPath() string { return filepath.Join(m.Dir, "status.json") }
func (m Manager) disabledPath() string { return filepath.Join(m.Dir, "disabled") }
func (m Manager) lockPath() string     { return filepath.Join(m.Dir, "refresh.lock") }

func (m Manager) acquire(ctx context.Context) (func(), error) {
	if err := os.MkdirAll(m.Dir, 0o700); err != nil {
		return nil, err
	}
	return lockFile(ctx, m.lockPath())
}

func (m Manager) Refresh(ctx context.Context, data []byte, metadata Metadata) error {
	if _, err := os.Stat(m.disabledPath()); err == nil {
		return errors.New("adapter is disabled")
	}
	release, err := m.acquire(ctx)
	if err != nil {
		return err
	}
	defer release()
	if err := os.MkdirAll(m.Dir, 0o700); err != nil {
		return err
	}
	if _, err := os.Stat(m.providerPath()); err == nil {
		if err := copyFile(m.providerPath(), m.previousPath(), 0o600); err != nil {
			return fmt.Errorf("backup current provider: %w", err)
		}
	}
	if err := atomicWrite(m.providerPath(), data, 0o600); err != nil {
		return err
	}
	if metadata.Digest == "" {
		sum := sha256.Sum256(data)
		metadata.Digest = hex.EncodeToString(sum[:])
	}
	status := Status{Enabled: true, Provider: filepath.Base(m.providerPath()), NodeCount: metadata.NodeCount, Digest: metadata.Digest, LastSuccess: time.Now().UTC()}
	return writeStatus(m.metadataPath(), status)
}

// RecordFailure stores only a non-sensitive category and timestamp. It is
// safe to call after any failed stage; the current provider is untouched.
func (m Manager) RecordFailure(ctx context.Context, class string) error {
	release, err := m.acquire(ctx)
	if err != nil {
		return err
	}
	defer release()
	status, err := m.Status()
	if err != nil {
		return err
	}
	status.LastFailure = time.Now().UTC()
	status.FailureClass = class
	return writeStatus(m.metadataPath(), status)
}

func (m Manager) Rollback(ctx context.Context) error {
	release, err := m.acquire(ctx)
	if err != nil {
		return err
	}
	defer release()
	if _, err := os.Stat(m.previousPath()); err != nil {
		return fmt.Errorf("no previous provider: %w", err)
	}
	if err := os.Rename(m.previousPath(), m.providerPath()); err != nil {
		return fmt.Errorf("restore previous provider: %w", err)
	}
	return nil
}

func (m Manager) Disable(ctx context.Context) error {
	release, err := m.acquire(ctx)
	if err != nil {
		return err
	}
	defer release()
	if err := os.WriteFile(m.disabledPath(), []byte("disabled\n"), 0o600); err != nil {
		return err
	}
	return nil
}

func (m Manager) Enable(ctx context.Context) error {
	release, err := m.acquire(ctx)
	if err != nil {
		return err
	}
	defer release()
	err = os.Remove(m.disabledPath())
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (m Manager) Status() (Status, error) {
	var status Status
	if _, err := os.Stat(m.disabledPath()); err == nil {
		status.Enabled = false
	} else {
		status.Enabled = true
	}
	b, err := os.ReadFile(m.metadataPath())
	if os.IsNotExist(err) {
		status.Provider = filepath.Base(m.providerPath())
		return status, nil
	}
	if err != nil {
		return status, err
	}
	if err := json.Unmarshal(b, &status); err != nil {
		return status, fmt.Errorf("read status: %w", err)
	}
	if _, err := os.Stat(m.disabledPath()); err == nil {
		status.Enabled = false
	}
	return status, nil
}

func atomicWrite(path string, data []byte, mode os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".provider-*")
	if err != nil {
		return fmt.Errorf("create candidate: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(mode); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write candidate: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("replace provider: %w", err)
	}
	return nil
}

func copyFile(src, dst string, mode os.FileMode) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return atomicWrite(dst, b, mode)
}

func writeStatus(path string, status Status) error {
	b, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return err
	}
	return atomicWrite(path, append(b, '\n'), 0o600)
}
