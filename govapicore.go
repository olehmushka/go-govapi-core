// Copyright 2026 Oleh Mushka
// SPDX-License-Identifier: Apache-2.0

// Package govapicore is the shared HTTP-client kernel for a family of small Go clients to
// public/government APIs (go-interpol-client, go-factbook-client, go-wof-client,
// go-wikidata-client, ...). It exists because each of those hand-rolled the same
// *http.Client construction, User-Agent resolution, and bounded JSON-GET boilerplate —
// extracted from go-oikumenea's internal/hermenea/fetcher package, which had the
// identical pattern copy-pasted across five separate fetchers.
//
// Deliberately minimal: only what those clients actually share today. Batch-import
// concerns like source-version/checksum staging stay in hermenea, and streaming a large
// compressed download to disk (needed only by the WOF gazetteer client) stays in
// go-wof-client rather than living here unused by everyone else.
package govapicore

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// DefaultMaxBody bounds a GetJSON response body (16 MiB) when maxBytes <= 0, so a
// runaway or hostile endpoint can't exhaust memory.
const DefaultMaxBody int64 = 16 << 20

// NewHTTPClient returns an *http.Client bounded by timeout. timeout <= 0 means
// unbounded — the caller's context.Context deadline governs instead. This mirrors
// go-oikumenea's hermenea fetchers: sources that stream large payloads (a planet-scale
// WOF distribution, the Glottolog CLDF values.csv) intentionally use no fixed client
// deadline and rely on a job-level context instead, while a small point-lookup client
// wants a real bound.
func NewHTTPClient(timeout time.Duration) *http.Client {
	if timeout <= 0 {
		return &http.Client{}
	}
	return &http.Client{Timeout: timeout}
}

// ResolveUserAgent returns the value of envVar if set and non-blank, else fallback.
// Several public APIs — the Wikidata SPARQL endpoint, iso639-3.sil.org — 403 a request
// carrying no identifying User-Agent or the bare Go default ("Go-http-client"), so a
// caller-identifying default is required, not optional; an operator embedding one of
// these clients in their own product can override it with their own contact via
// envVar without a code change.
func ResolveUserAgent(envVar, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(envVar)); v != "" {
		return v
	}
	return fallback
}

// ErrUnexpectedStatus is returned by GetJSON when the response status is not 200 OK.
// Body is truncated to a small preview so an error message can't itself become an
// unbounded read.
type ErrUnexpectedStatus struct {
	URL    string
	Status string
	Body   []byte
}

func (e *ErrUnexpectedStatus) Error() string {
	return fmt.Sprintf("govapicore: GET %s: unexpected status %s: %s", e.URL, e.Status, e.Body)
}

// errStatusPreview bounds how much of a non-200 response body ErrUnexpectedStatus
// carries — enough to diagnose, small enough to never be the unbounded read itself.
const errStatusPreview = 2048

// GetJSON issues a GET request to url with the given headers, reads at most maxBytes of
// the response body (maxBytes <= 0 uses DefaultMaxBody), and decodes it as JSON into
// out. A non-200 response is reported as *ErrUnexpectedStatus, not a decode error, so a
// caller can distinguish "upstream is down" from "upstream sent something this client
// doesn't understand".
func GetJSON(ctx context.Context, client *http.Client, url string, headers map[string]string, maxBytes int64, out any) error {
	if maxBytes <= 0 {
		maxBytes = DefaultMaxBody
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("govapicore: build request for %s: %w", url, err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("govapicore: GET %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		preview, _ := io.ReadAll(io.LimitReader(resp.Body, errStatusPreview))
		return &ErrUnexpectedStatus{URL: url, Status: resp.Status, Body: preview}
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes))
	if err != nil {
		return fmt.Errorf("govapicore: read response from %s: %w", url, err)
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("govapicore: decode response from %s: %w", url, err)
	}
	return nil
}
