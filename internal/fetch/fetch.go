// Package fetch downloads encrypted subscriptions with bounded resources.
package fetch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	DefaultMaxBytes = 16 << 20
	DefaultTimeout  = 45 * time.Second
)

var ErrMissingEncryptionMarker = errors.New("subscription-encryption header is missing or false")
var ErrResponseTooLarge = errors.New("subscription response exceeds size limit")
var ErrInsecureURL = errors.New("subscription URL must use HTTPS")

type Client struct {
	HTTPClient *http.Client
	MaxBytes   int64
}

type Response struct {
	Body        []byte
	StatusCode  int
	ContentType string
	Encrypted   bool
}

func (c Client) Do(ctx context.Context, rawURL string) (Response, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return Response{}, ErrInsecureURL
	}
	hc := c.HTTPClient
	if hc == nil {
		hc = http.DefaultClient
	}
	max := c.MaxBytes
	if max <= 0 {
		max = DefaultMaxBytes
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return Response{}, errors.New("invalid subscription request")
	}
	resp, err := hc.Do(req)
	if err != nil {
		return Response{}, errors.New("download subscription request failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return Response{StatusCode: resp.StatusCode, ContentType: resp.Header.Get("Content-Type")}, fmt.Errorf("subscription returned HTTP %d", resp.StatusCode)
	}
	limited := io.LimitReader(resp.Body, max+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return Response{}, errors.New("read subscription response failed")
	}
	if int64(len(body)) > max {
		return Response{}, ErrResponseTooLarge
	}
	marker := strings.EqualFold(strings.TrimSpace(resp.Header.Get("subscription-encryption")), "true")
	if !marker {
		return Response{Body: body, StatusCode: resp.StatusCode, ContentType: resp.Header.Get("Content-Type")}, ErrMissingEncryptionMarker
	}
	return Response{Body: body, StatusCode: resp.StatusCode, ContentType: resp.Header.Get("Content-Type"), Encrypted: true}, nil
}
