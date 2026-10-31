package credentials

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileStoreProtectsAndRoundTrips(t *testing.T) {
	dir := t.TempDir()
	s := FileStore{Dir: dir}
	if err := s.Set("password", "fixture-secret"); err != nil {
		t.Fatal(err)
	}
	got, err := s.Get("password")
	if err != nil || got != "fixture-secret" {
		t.Fatalf("got %q, %v", got, err)
	}
	info, err := os.Stat(filepath.Join(dir, "password"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode %o", info.Mode().Perm())
	}
}
