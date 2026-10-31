package fetch

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// The transport is in-process so the tests also run in locked-down CI where
// binding a loopback socket is unavailable. httptest.NewRequest still checks
// the request shape supplied to the transport.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func fixtureResponse(body string, headers http.Header) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     headers,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
		Request:    httptest.NewRequest(http.MethodGet, "https://fixture.invalid/subscription", nil),
	}
}

func TestDoChecksMarkerAndBounds(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return fixtureResponse("fixture", http.Header{"Subscription-Encryption": []string{"true"}}), nil
	})
	client := &http.Client{Transport: transport}
	res, err := (Client{HTTPClient: client, MaxBytes: 100}).Do(context.Background(), "https://fixture.invalid/subscription")
	if err != nil {
		t.Fatal(err)
	}
	if !res.Encrypted || string(res.Body) != "fixture" {
		t.Fatalf("unexpected response: %+v", res)
	}

	tooSmall := Client{HTTPClient: client, MaxBytes: 3}
	if _, err := tooSmall.Do(context.Background(), "https://fixture.invalid/subscription"); err != ErrResponseTooLarge {
		t.Fatalf("got %v, want size error", err)
	}
}

func TestDoRejectsMissingMarker(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return fixtureResponse("fixture", nil), nil
	})
	_, err := (Client{HTTPClient: &http.Client{Transport: transport}}).Do(context.Background(), "https://fixture.invalid/subscription")
	if err != ErrMissingEncryptionMarker {
		t.Fatalf("got %v", err)
	}
}

func TestDoRejectsNonHTTPSURL(t *testing.T) {
	_, err := (Client{}).Do(context.Background(), "http://fixture.invalid/subscription")
	if err != ErrInsecureURL {
		t.Fatalf("got %v, want %v", err, ErrInsecureURL)
	}
}
