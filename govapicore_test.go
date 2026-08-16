// Copyright 2026 Oleh Mushka
// SPDX-License-Identifier: Apache-2.0

package govapicore

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type payload struct {
	Name string `json:"name"`
}

func TestGetJSONSuccess(t *testing.T) {
	var gotHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"villa maria"}`))
	}))
	defer srv.Close()

	client := NewHTTPClient(5 * time.Second)
	var out payload
	err := GetJSON(context.Background(), client, srv.URL, map[string]string{"User-Agent": "test-agent/1.0"}, 0, &out)
	if err != nil {
		t.Fatalf("GetJSON: %v", err)
	}
	if out.Name != "villa maria" {
		t.Fatalf("out.Name = %q, want %q", out.Name, "villa maria")
	}
	if gotHeader != "test-agent/1.0" {
		t.Fatalf("User-Agent = %q, want %q — headers were not forwarded", gotHeader, "test-agent/1.0")
	}
}

func TestGetJSONUnexpectedStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("upstream is down"))
	}))
	defer srv.Close()

	client := NewHTTPClient(5 * time.Second)
	var out payload
	err := GetJSON(context.Background(), client, srv.URL, nil, 0, &out)
	if err == nil {
		t.Fatal("a real 500 from the upstream should be a real error")
	}
	var statusErr *ErrUnexpectedStatus
	if !errors.As(err, &statusErr) {
		t.Fatalf("err = %v (%T), want *ErrUnexpectedStatus — a caller needs to distinguish 'upstream is down' from 'upstream sent something unparseable'", err, err)
	}
	if statusErr.Status != "500 Internal Server Error" {
		t.Fatalf("Status = %q, want %q", statusErr.Status, "500 Internal Server Error")
	}
	if string(statusErr.Body) != "upstream is down" {
		t.Fatalf("Body = %q, want %q", statusErr.Body, "upstream is down")
	}
}

func TestGetJSONBodyCappedAtMaxBytes(t *testing.T) {
	// A response larger than maxBytes must not be decoded as if it were complete — the
	// cap exists to bound memory against a hostile/misbehaving endpoint, not to
	// silently truncate a valid response into something that happens to still parse.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"` + string(make([]byte, 100)) + `-truncated-me"}`))
	}))
	defer srv.Close()

	client := NewHTTPClient(5 * time.Second)
	var out payload
	err := GetJSON(context.Background(), client, srv.URL, nil, 8, &out)
	if err == nil {
		t.Fatal("a body larger than maxBytes should fail to decode as truncated JSON, not silently succeed")
	}
}

func TestResolveUserAgentEnvOverride(t *testing.T) {
	t.Setenv("GOVAPICORE_TEST_UA", "myapp/1.0 (contact@example.com)")
	got := ResolveUserAgent("GOVAPICORE_TEST_UA", "fallback/1.0")
	if got != "myapp/1.0 (contact@example.com)" {
		t.Fatalf("ResolveUserAgent = %q, want the env override", got)
	}
}

func TestResolveUserAgentFallback(t *testing.T) {
	t.Setenv("GOVAPICORE_TEST_UA_UNSET", "")
	got := ResolveUserAgent("GOVAPICORE_TEST_UA_UNSET", "fallback/1.0")
	if got != "fallback/1.0" {
		t.Fatalf("ResolveUserAgent = %q, want the fallback when the env var is blank", got)
	}
}

func TestNewHTTPClientTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewHTTPClient(10 * time.Millisecond)
	var out payload
	err := GetJSON(context.Background(), client, srv.URL, nil, 0, &out)
	if err == nil {
		t.Fatal("a client bounded to 10ms against a 100ms-slow handler should time out, not succeed")
	}
}

func TestNewHTTPClientUnboundedRespectsContextDeadline(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewHTTPClient(0) // unbounded: the context deadline below must still govern
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	var out payload
	err := GetJSON(ctx, client, srv.URL, nil, 0, &out)
	if err == nil {
		t.Fatal("an unbounded client (timeout<=0) should still respect the caller's context deadline")
	}
}
