# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-08-16

### Added

- Initial extraction from go-oikumenea's `internal/hermenea/fetcher`: `NewHTTPClient`,
  `ResolveUserAgent`, `GetJSON`, `ErrUnexpectedStatus` — the shared HTTP-client kernel for the
  go-interpol-client / go-factbook-client / go-wof-client / go-wikidata-client family.

[Unreleased]: https://github.com/olehmushka/go-govapi-core/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/olehmushka/go-govapi-core/releases/tag/v0.1.0
