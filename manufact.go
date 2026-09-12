package canimcp

import "strings"

// clientNameManufact is the product token the Manufact Cloud MCP client
// (manufact.com, built on mcp-use) sends in clientInfo.name, in both the
// initialize params and the per-request _meta. Manufact Cloud is a remote
// web-dashboard client: its User-Agent is a plain Chrome browser string
// shared with millions of HTTP clients, so clientInfo.name is the only
// reliable identity signal.
const clientNameManufact = "manufact cloud"

// manufactDetector matches the Manufact Cloud remote HTTP client.
// Its capability surface is the generic HTTP profile plus MCP Apps UI and
// form/URL elicitation, both negotiated on the wire (initialize
// capabilities: io.modelcontextprotocol/ui extension with
// text/html;profile=mcp-app, and elicitation form+url). It is its own
// HostType so callers can gate on HostIs(HostManufact).
type manufactDetector struct{}

func (manufactDetector) Match(req Evidence) (HostType, AuthMethod) {
	// Manufact Cloud is a remote HTTP dashboard client — never co-located
	// and never on the OpenAI-specific tunnel transport.
	if req.CoLocated || req.TunnelOpenAI {
		return HostUnknown, ""
	}
	if req.ClientInfo != nil && strings.EqualFold(strings.TrimSpace(req.ClientInfo.Name), clientNameManufact) {
		return HostManufact, authFromToken(req.TokenInfo)
	}
	return HostUnknown, ""
}
