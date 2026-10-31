package runtime

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"
)

func TestRefreshRollbackAndDisable(t *testing.T) {
	m := Manager{Dir: t.TempDir()}
	ctx := context.Background()
	if err := m.Refresh(ctx, []byte("proxies: one\n"), Metadata{NodeCount: 1}); err != nil {
		t.Fatal(err)
	}
	if err := m.Refresh(ctx, []byte("proxies: two\n"), Metadata{NodeCount: 2}); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(m.providerPath())
	if err != nil || string(b) != "proxies: two\n" {
		t.Fatalf("current provider: %q, %v", b, err)
	}
	if err := m.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	b, err = os.ReadFile(m.providerPath())
	if err != nil || string(b) != "proxies: one\n" {
		t.Fatalf("rolled back provider: %q, %v", b, err)
	}
	if err := m.Disable(ctx); err != nil {
		t.Fatal(err)
	}
	if err := m.Refresh(ctx, []byte("proxies: three\n"), Metadata{}); err == nil || !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("got %v", err)
	}
	if err := m.Enable(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestRefreshSerializesConcurrentWriters(t *testing.T) {
	m := Manager{Dir: t.TempDir()}
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for _, value := range []string{"one", "two"} {
		value := value
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- m.Refresh(context.Background(), []byte(value), Metadata{})
		}()
	}
	wg.Wait()
	close(results)
	for err := range results {
		if err != nil {
			t.Fatal(err)
		}
	}
	b, err := os.ReadFile(m.providerPath())
	if err != nil || (string(b) != "one" && string(b) != "two") {
		t.Fatalf("unexpected final provider %q: %v", b, err)
	}
}

func TestRefreshDoesNotLeavePartialCandidate(t *testing.T) {
	m := Manager{Dir: t.TempDir()}
	if err := m.Refresh(context.Background(), []byte("stable"), Metadata{}); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(m.Dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".provider-") {
			t.Fatalf("temporary candidate left behind: %s", entry.Name())
		}
	}
}
