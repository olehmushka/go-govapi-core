# Security Policy

## Supported versions

Only the latest tagged release is supported. Please upgrade before reporting an issue that may
already be fixed.

## Reporting a vulnerability

Please do **not** open a public GitHub issue for security vulnerabilities.

- **Preferred**: use GitHub's private vulnerability reporting for this repository
  (the "Security" tab → "Report a vulnerability").
- **Alternative**: email olegamysk@gmail.com with details and, if possible, a proof of concept.

You should get an initial response within a few days.

## Scope notes

This package builds `*http.Client`s and issues GET requests to caller-supplied URLs, decoding
caller-supplied JSON responses into caller-supplied Go types. It does not execute untrusted input
beyond JSON decoding, and holds no credentials/secrets itself. The most likely real-world issue
class here is a resource-exhaustion bug (e.g. an unbounded read from a malicious or compromised
endpoint) — `GetJSON`'s `maxBytes` cap exists specifically to bound that, not as a formality.
