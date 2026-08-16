# go-govapi-core

[![CI](https://github.com/olehmushka/go-govapi-core/actions/workflows/ci.yml/badge.svg)](https://github.com/olehmushka/go-govapi-core/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/olehmushka/go-govapi-core.svg)](https://pkg.go.dev/github.com/olehmushka/go-govapi-core)
[![Go Report Card](https://goreportcard.com/badge/github.com/olehmushka/go-govapi-core)](https://goreportcard.com/report/github.com/olehmushka/go-govapi-core)
[![License](https://img.shields.io/github/license/olehmushka/go-govapi-core)](LICENSE)
[![Latest Release](https://img.shields.io/github/v/tag/olehmushka/go-govapi-core)](https://github.com/olehmushka/go-govapi-core/releases)

The shared HTTP-client kernel for a family of small Go clients to public/government
APIs — [go-interpol-client](https://github.com/olehmushka/go-interpol-client),
[go-factbook-client](https://github.com/olehmushka/go-factbook-client),
[go-wof-client](https://github.com/olehmushka/go-wof-client),
[go-wikidata-client](https://github.com/olehmushka/go-wikidata-client). Zero
third-party dependencies.

## Install

```sh
go get github.com/olehmushka/go-govapi-core
```

## Usage

```go
client := govapicore.NewHTTPClient(10 * time.Second)
ua := govapicore.ResolveUserAgent("MYAPP_USER_AGENT", "myapp/1.0 (contact@example.com)")

var out MyResponseType
err := govapicore.GetJSON(ctx, client, url, map[string]string{"User-Agent": ua}, 0, &out)
if err != nil {
    var statusErr *govapicore.ErrUnexpectedStatus
    if errors.As(err, &statusErr) {
        // the upstream API itself returned a non-200 — statusErr.Status / .Body
    }
    // otherwise: a network error or a response this client couldn't decode
}
```

## Why this exists

Each of the client packages above talks to a different upstream API but was hand-rolling
the exact same three things: a bounded `*http.Client`, an env-overridable User-Agent
(several public APIs — Wikidata's SPARQL endpoint among them — 403 a request with no
identifying UA), and a GET-JSON-with-a-size-cap-and-status-check helper. This package
extracts just that shared shape — no more — from
[go-oikumenea](https://github.com/olehmushka/go-oikumenea)'s
`internal/hermenea/fetcher` package, which had the identical pattern copy-pasted across
five separate fetchers before this repo existed.

It is deliberately **not** a general-purpose HTTP framework: no retries, no caching, no
batch/versioning concepts. A caller-supplied `*http.Client` and `context.Context` still
govern the real transport behavior; this package only removes the boilerplate around
them.

## License

Apache 2.0 — see [LICENSE](./LICENSE).
