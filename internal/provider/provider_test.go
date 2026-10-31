package provider

import (
	"strings"
	"testing"
)

const fixture = `proxies:
  - name: one
    type: http
    server: example.invalid
    port: 443
  - name: two
    type: socks5
    server: example.invalid
    port: 1080
rules:
  - MATCH,DIRECT
`

func TestNormalizeExtractsProxies(t *testing.T) {
	result, err := Normalize([]byte(fixture))
	if err != nil {
		t.Fatal(err)
	}
	if result.NodeCount != 2 || result.Digest == "" {
		t.Fatalf("unexpected result: %+v", result)
	}
	if !strings.Contains(string(result.YAML), "name: one") || strings.Contains(string(result.YAML), "rules:") {
		t.Fatalf("provider was not minimized: %s", result.YAML)
	}
}

func TestNormalizeRejectsInvalidProfiles(t *testing.T) {
	tests := []struct {
		name, input string
		want        error
	}{
		{"missing", "dns: {}\n", ErrMissingProxies},
		{"empty", "proxies: []\n", ErrMissingProxies},
		{"duplicate", "proxies:\n - {name: same, type: http}\n - {name: same, type: http}\n", ErrDuplicateName},
		{"empty name", "proxies:\n - {name: '  ', type: http}\n", ErrInvalidProxy},
		{"missing type", "proxies:\n - {name: one}\n", ErrInvalidProxy},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Normalize([]byte(test.input))
			if err == nil || !strings.Contains(err.Error(), test.want.Error()) {
				t.Fatalf("got %v, want %v", err, test.want)
			}
		})
	}
}
