# canimcp

[![Go Version](https://img.shields.io/badge/Go-1.26.0-blue)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Build Status](https://github.com/LumeWeb/canimcp/actions/workflows/go.yml/badge.svg)](https://github.com/LumeWeb/canimcp/actions/workflows/go.yml)

A caniuse-style compatibility database and detector for MCP hosts,
transports, and wire capabilities.

An MCP server asks one question: which capabilities does the connected host
on the wire actually support? `canimcp` models it like caniuse — a static
`Profile` declares which features a host + transport combination supports,
overlaid with runtime wire signals (clientInfo, headers, auth) resolved by
ordered `Detector`s. No detector match degrades to a generic profile for the
transport.

The core package imports no MCP SDK and carries no terminal UI or product
concerns; deployment/domain policies belong to the consuming application.

## Example usage

```go
reg := canimcp.NewRegistry()

prof := reg.DetectFromHTTPRequest(header, true, false, nil)
if prof.Has(canimcp.FeatMCPApps) {
    // the host renders MCP Apps UI
}
if canimcp.HostIs(canimcp.HostGrok)(prof) {
    // host-specific behavior as a predicate over Profile
}
```

## License

MIT
