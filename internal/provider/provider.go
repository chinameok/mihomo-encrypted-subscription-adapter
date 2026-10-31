// Package provider validates decrypted Clash/Mihomo profiles and emits the
// minimal file-provider document consumed by Mihomo.
package provider

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	ErrMissingProxies = errors.New("profile does not contain proxies")
	ErrInvalidProxy   = errors.New("profile contains an invalid proxy")
	ErrDuplicateName  = errors.New("profile contains duplicate proxy name")
)

type Result struct {
	YAML      []byte
	NodeCount int
	Digest    string
}

type profile struct {
	Proxies []map[string]any `yaml:"proxies"`
}

func Normalize(input []byte) (Result, error) {
	var p profile
	if err := yaml.Unmarshal(input, &p); err != nil {
		return Result{}, fmt.Errorf("parse profile YAML: %w", err)
	}
	if p.Proxies == nil || len(p.Proxies) == 0 {
		return Result{}, ErrMissingProxies
	}
	seen := make(map[string]struct{}, len(p.Proxies))
	for i, proxy := range p.Proxies {
		if proxy == nil {
			return Result{}, fmt.Errorf("%w at index %d", ErrInvalidProxy, i)
		}
		nameValue, ok := proxy["name"]
		name, nameOK := nameValue.(string)
		if !ok || !nameOK || strings.TrimSpace(name) == "" {
			return Result{}, fmt.Errorf("%w at index %d: name is required", ErrInvalidProxy, i)
		}
		proxyType, typeOK := proxy["type"].(string)
		if !typeOK || strings.TrimSpace(proxyType) == "" {
			return Result{}, fmt.Errorf("%w at index %d: type is required", ErrInvalidProxy, i)
		}
		name = strings.TrimSpace(name)
		if _, exists := seen[name]; exists {
			return Result{}, fmt.Errorf("%w: %q", ErrDuplicateName, name)
		}
		seen[name] = struct{}{}
	}
	output, err := yaml.Marshal(struct {
		Proxies []map[string]any `yaml:"proxies"`
	}{Proxies: p.Proxies})
	if err != nil {
		return Result{}, fmt.Errorf("marshal provider: %w", err)
	}
	digest := sha256.Sum256(output)
	return Result{YAML: output, NodeCount: len(p.Proxies), Digest: hex.EncodeToString(digest[:])}, nil
}
