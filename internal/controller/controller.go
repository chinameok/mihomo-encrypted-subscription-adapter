// Package controller contains the narrow Mihomo controller operation needed
// after an atomic provider update. It is deliberately optional: a failed
// controller refresh never causes the adapter to restart Mihomo implicitly.
package controller

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	BaseURL string
	Secret  string
	HTTP    *http.Client
}

// RefreshProvider asks Mihomo to reload a file provider. The endpoint is
// configurable because vendor forks can expose a different controller path.
func (c Client) RefreshProvider(ctx context.Context, endpoint, name string) error {
	if endpoint == "" {
		endpoint = strings.TrimRight(c.BaseURL, "/") + "/providers/proxies/" + url.PathEscape(name)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewReader(nil))
	if err != nil {
		return fmt.Errorf("create Mihomo refresh request: %w", err)
	}
	if c.Secret != "" {
		req.Header.Set("Authorization", "Bearer "+c.Secret)
	}
	hc := c.HTTP
	if hc == nil {
		hc = http.DefaultClient
	}
	resp, err := hc.Do(req)
	if err != nil {
		return fmt.Errorf("refresh Mihomo provider: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("Mihomo provider refresh returned HTTP %d", resp.StatusCode)
	}
	return nil
}
