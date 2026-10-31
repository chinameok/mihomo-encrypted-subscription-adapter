package controller

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestRefreshProviderUsesBearerAndEscapesName(t *testing.T) {
	var got *http.Request
	client := Client{
		BaseURL: "http://127.0.0.1:19090",
		Secret:  "fixture-secret",
		HTTP: &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
			got = r
			return &http.Response{StatusCode: http.StatusNoContent, Body: io.NopCloser(bytes.NewReader(nil)), Header: make(http.Header), Request: r}, nil
		})},
	}
	if err := client.RefreshProvider(context.Background(), "", "provider/fixture"); err != nil {
		t.Fatal(err)
	}
	if got.Method != http.MethodPut || got.URL.Path != "/providers/proxies/provider/fixture" || got.Header.Get("Authorization") != "Bearer fixture-secret" {
		t.Fatalf("unexpected request: %s %s auth=%q", got.Method, got.URL.Path, got.Header.Get("Authorization"))
	}
}
