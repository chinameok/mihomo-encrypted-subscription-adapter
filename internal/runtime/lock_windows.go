//go:build windows

package runtime

import (
	"context"
	"fmt"
	"os"
	"time"
)

// Windows has no syscall.Flock equivalent in the standard library. Exclusive
// creation gives launchd-equivalent single-writer behavior for this CLI; a
// stale file is harmlessly retried by the operator after checking the process.
func lockFile(ctx context.Context, path string) (func(), error) {
	for {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err == nil {
			return func() { _ = f.Close(); _ = os.Remove(path) }, nil
		}
		if !os.IsExist(err) {
			return nil, fmt.Errorf("open refresh lock: %w", err)
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(25 * time.Millisecond):
		}
	}
}
