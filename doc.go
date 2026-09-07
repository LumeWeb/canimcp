// Package canimcp is a caniuse-style compatibility database and detector
// for MCP hosts, transports, and wire capabilities.
//
// The package answers one question for an MCP server: which capabilities
// does the connected host on the wire actually support? A caniuse analogy
// drives the model — a static [Profile] declares which features a
// HostType + Transport combination supports, overlaid with runtime wire
// signals (clientInfo, headers, auth) resolved by ordered detectors.
package canimcp
