package canimcp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func newTokenInfo() *TokenInfo {
	return &TokenInfo{
		Scopes:     []string{"read", "write"},
		Expiration: time.Date(2025, time.December, 31, 23, 59, 59, 0, time.UTC),
		UserID:     "user-123",
		Extra: map[string]any{
			"iss": "test-issuer",
		},
	}
}

// TestAndPredicate verifies the And combinator passes only when every given
// predicate passes, including with zero predicates (trivially true).

func TestAndPredicate(t *testing.T) {
	// Profiles built inline so the predicate model is pinned independently
	// of the static per-host profiles.
	tunnel := Profile{HostType: HostGrok, Transport: TransportOpenAI}
	http := Profile{HostType: HostGrok, Transport: TransportHTTP}

	isGrok := HostIs(HostGrok)
	isHTTP := TransportIs(TransportHTTP)
	require.True(t, And()(http), "no predicates must be trivially true")
	require.True(t, isGrok(http))
	require.False(t, And(isGrok, isHTTP)(tunnel), "tunnel profile must fail the http conjunct")
	require.True(t, And(isGrok, isHTTP)(http), "http grok must match both conjuncts")
	require.False(t, And(isGrok, Not(isHTTP))(http), "http grok must not match the !http conjunct")
}

// ---------------------------------------------------------------------------
// Detector tests
// ---------------------------------------------------------------------------

func TestResolveProfile(t *testing.T) {
	tests := []struct {
		name      string
		host      HostType
		transport TransportKind
		auth      AuthMethod
		want      Profile
	}{
		{
			name:      "openai over openai tunnel",
			host:      HostOpenAI,
			transport: TransportOpenAI,
			auth:      AuthNone,
			want:      ProfileOpenAITunnel,
		},
		{
			name:      "chatgpt over openai tunnel",
			host:      HostChatGPT,
			transport: TransportOpenAI,
			auth:      AuthNone,
			want:      ProfileOpenAITunnel,
		},
		{
			name:      "openai over http",
			host:      HostOpenAI,
			transport: TransportHTTP,
			auth:      AuthOAuth,
			want:      ProfileOpenAIHTTP,
		},
		{
			name:      "chatgpt over http",
			host:      HostChatGPT,
			transport: TransportHTTP,
			auth:      AuthOAuth,
			want:      ProfileOpenAIHTTP,
		},
		{
			name:      "grok over http",
			host:      HostGrok,
			transport: TransportHTTP,
			auth:      AuthOAuth,
			want:      ProfileGrokHTTP,
		},
		{
			name:      "grok over stdio",
			host:      HostGrok,
			transport: TransportStdio,
			auth:      AuthNone,
			want:      ProfileGrokStdio,
		},
		{
			name:      "claude over http",
			host:      HostClaude,
			transport: TransportHTTP,
			auth:      AuthOAuth,
			want:      ProfileClaudeHTTP,
		},
		{
			name:      "generic over stdio",
			host:      HostGeneric,
			transport: TransportStdio,
			auth:      AuthNone,
			want:      ProfileStdioGeneric,
		},
		{
			name:      "generic over http",
			host:      HostGeneric,
			transport: TransportHTTP,
			auth:      AuthBearer,
			want:      ProfileHTTPGeneric,
		},
		{
			name:      "unknown over stdio falls to stdio generic",
			host:      HostUnknown,
			transport: TransportStdio,
			auth:      AuthNone,
			want:      ProfileStdioGeneric,
		},
		{
			name:      "unknown over http falls to http generic",
			host:      HostUnknown,
			transport: TransportHTTP,
			auth:      AuthBearer,
			want:      ProfileHTTPGeneric,
		},
		{
			name:      "unknown combo falls back to stdio generic",
			host:      HostUnknown,
			transport: TransportKind("bogus"),
			auth:      AuthNone,
			want:      ProfileStdioGeneric,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveProfile(tt.host, tt.transport, tt.auth)
			require.Equal(t, tt.want.HostType, got.HostType)
			require.Equal(t, tt.want.Transport, got.Transport)
			require.Equal(t, tt.want.AuthMethod, got.AuthMethod)
			require.Equal(t, tt.want.Remote, got.Remote)
			require.Equal(t, tt.want.Features, got.Features)
		})
	}
}

// TestResolveProfile_AiderDesk verifies the alias path: an aliased host
// inherits the target's full profile (features, transport, auth, remote) but
// keeps its own HostType.

func TestResolveProfile_AiderDesk(t *testing.T) {
	got := resolveProfile(HostAiderDesk, TransportStdio, AuthNone)

	require.Equal(t, HostAiderDesk, got.HostType)
	require.Equal(t, TransportStdio, got.Transport)
	require.Equal(t, AuthNone, got.AuthMethod)
	require.False(t, got.Remote)
	// Feature surface equals the generic stdio profile exactly (alias target).
	require.Equal(t, ProfileStdioGeneric.Features, got.Features)
}

// TestResolveProfile_Goose verifies the Goose alias path: an aliased host
// inherits the target's full profile (features, transport, auth, remote) but
// keeps its own HostType.

func TestResolveProfile_Goose(t *testing.T) {
	got := resolveProfile(HostGoose, TransportStdio, AuthNone)

	require.Equal(t, HostGoose, got.HostType)
	require.Equal(t, TransportStdio, got.Transport)
	require.Equal(t, AuthNone, got.AuthMethod)
	require.False(t, got.Remote)
	// Feature surface equals the shared stdio+MCP-Apps profile exactly (alias target).
	require.Equal(t, ProfileStdioMCPApps.Features, got.Features)
}

// TestResolveProfile_ClaudeDesktop verifies the Claude Desktop alias path: it
// shares the same generic stdio+MCP-Apps declaration as Goose (both target
// HostStdioApps) but keeps its own HostType.

func TestResolveProfile_ClaudeDesktop(t *testing.T) {
	got := resolveProfile(HostClaudeDesktop, TransportStdio, AuthNone)

	require.Equal(t, HostClaudeDesktop, got.HostType)
	require.Equal(t, TransportStdio, got.Transport)
	require.Equal(t, AuthNone, got.AuthMethod)
	require.False(t, got.Remote)
	// Both Claude Desktop and Goose resolve to the same shared declaration.
	require.Equal(t, ProfileStdioMCPApps.Features, got.Features)
	require.Equal(t, resolveProfile(HostGoose, TransportStdio, AuthNone).Features, got.Features)
}

// TestResolveProfile_ClaudeCode verifies the Claude Code alias path: it shares
// the same generic stdio+MCP-Apps declaration as Claude Desktop and Goose (all
// target HostStdioApps) but keeps its own HostType.

func TestResolveProfile_ClaudeCode(t *testing.T) {
	got := resolveProfile(HostClaudeCode, TransportStdio, AuthNone)

	require.Equal(t, HostClaudeCode, got.HostType)
	require.Equal(t, TransportStdio, got.Transport)
	require.Equal(t, AuthNone, got.AuthMethod)
	require.False(t, got.Remote)
	// Resolves to the same shared stdio+MCP-Apps declaration as its peers.
	require.Equal(t, ProfileStdioMCPApps.Features, got.Features)
	require.True(t, got.Has(FeatMCPApps))
	require.Equal(t, resolveProfile(HostGoose, TransportStdio, AuthNone).Features, got.Features)
}

// TestResolveProfile_Devin verifies the Devin alias path: an aliased host
// inherits the target's full profile (features, transport, auth, remote) but
// keeps its own HostType.

func TestResolveProfile_Devin(t *testing.T) {
	got := resolveProfile(HostDevin, TransportStdio, AuthNone)

	require.Equal(t, HostDevin, got.HostType)
	require.Equal(t, TransportStdio, got.Transport)
	require.Equal(t, AuthNone, got.AuthMethod)
	require.False(t, got.Remote)
	// Feature surface equals the generic stdio profile exactly (alias target).
	require.Equal(t, ProfileStdioGeneric.Features, got.Features)
}

func TestResolveProfile_Codex(t *testing.T) {
	got := resolveProfile(HostCodex, TransportStdio, AuthNone)

	require.Equal(t, HostCodex, got.HostType)
	require.Equal(t, TransportStdio, got.Transport)
	require.Equal(t, AuthNone, got.AuthMethod)
	require.False(t, got.Remote)
	// Feature surface equals the generic stdio profile exactly (alias target).
	require.Equal(t, ProfileStdioGeneric.Features, got.Features)
}

// TestResolveProfile_CopilotCLI verifies the Copilot CLI alias path: an
// aliased host inherits the target's full profile (features, transport, auth,
// remote) but keeps its own HostType.

func TestResolveProfile_CopilotCLI(t *testing.T) {
	got := resolveProfile(HostCopilotCLI, TransportStdio, AuthNone)

	require.Equal(t, HostCopilotCLI, got.HostType)
	require.Equal(t, TransportStdio, got.Transport)
	require.Equal(t, AuthNone, got.AuthMethod)
	require.False(t, got.Remote)
	// Feature surface equals the generic stdio profile exactly (alias target).
	require.Equal(t, ProfileStdioGeneric.Features, got.Features)
}

// TestResolveProfile_OpenCode verifies the OpenCode alias path: an aliased
// host inherits the target's full profile (features, transport, auth, remote)
// but keeps its own HostType.

func TestResolveProfile_OpenCode(t *testing.T) {
	got := resolveProfile(HostOpenCode, TransportStdio, AuthNone)

	require.Equal(t, HostOpenCode, got.HostType)
	require.Equal(t, TransportStdio, got.Transport)
	require.Equal(t, AuthNone, got.AuthMethod)
	require.False(t, got.Remote)
	// Feature surface equals the generic stdio profile exactly (alias target).
	require.Equal(t, ProfileStdioGeneric.Features, got.Features)
}

// TestResolveProfile_Kimi verifies the Kimi alias path: an aliased host
// inherits the target's full profile (features, transport, auth, remote) but
// keeps its own HostType.

func TestResolveProfile_Kimi(t *testing.T) {
	got := resolveProfile(HostKimi, TransportStdio, AuthNone)

	require.Equal(t, HostKimi, got.HostType)
	require.Equal(t, TransportStdio, got.Transport)
	require.Equal(t, AuthNone, got.AuthMethod)
	require.False(t, got.Remote)
	// Feature surface equals the generic stdio profile exactly (alias target).
	require.Equal(t, ProfileStdioGeneric.Features, got.Features)
}

// TestResolveProfile_Fx verifies the fx alias path: an aliased host inherits
// the target's full profile (features, transport, auth, remote) but keeps its
// own HostType.

func TestResolveProfile_Fx(t *testing.T) {
	got := resolveProfile(HostFX, TransportStdio, AuthNone)

	require.Equal(t, HostFX, got.HostType)
	require.Equal(t, TransportStdio, got.Transport)
	require.Equal(t, AuthNone, got.AuthMethod)
	require.False(t, got.Remote)
	// Feature surface equals the generic stdio profile exactly (alias target).
	require.Equal(t, ProfileStdioGeneric.Features, got.Features)
	require.True(t, got.Has(FeatSinkDrop))
}

// TestProfileAliasDoesNotCorruptTarget locks in that resolving an alias
// returns a value copy — overriding the alias's HostType must not mutate the
// shared static target profile.

func TestProfileAliasDoesNotCorruptTarget(t *testing.T) {
	before := ProfileStdioGeneric.HostType

	_ = resolveProfile(HostAiderDesk, TransportStdio, AuthNone)

	require.Equal(t, before, ProfileStdioGeneric.HostType,
		"resolving an alias must not mutate the shared static target profile")
	require.Equal(t, HostGeneric, ProfileStdioGeneric.HostType)
}

// ---------------------------------------------------------------------------
// ProfileForTransport tests
// ---------------------------------------------------------------------------

func TestProfileForTransport(t *testing.T) {
	tests := []struct {
		name string
		t    TransportKind
		want Profile
	}{
		{name: "stdio", t: TransportStdio, want: ProfileStdioGeneric},
		{name: "http", t: TransportHTTP, want: ProfileHTTPGeneric},
		{name: "openai", t: TransportOpenAI, want: ProfileOpenAITunnel},
		{name: "unknown defaults to stdio", t: TransportKind("bogus"), want: ProfileStdioGeneric},
		{name: "empty defaults to stdio", t: TransportKind(""), want: ProfileStdioGeneric},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProfileForTransport(tt.t)
			require.Equal(t, tt.want.HostType, got.HostType)
			require.Equal(t, tt.want.Transport, got.Transport)
		})
	}
}

// ---------------------------------------------------------------------------
// FeatureSet tests
// ---------------------------------------------------------------------------

func TestFeatureSet_Has(t *testing.T) {
	fs := FeatureSet{
		FeatFileHostInput: true,
		FeatSourcePath:    true,
	}

	require.True(t, fs.Has(FeatFileHostInput))
	require.True(t, fs.Has(FeatSourcePath))
	require.False(t, fs.Has(FeatSourceMint))
	require.False(t, fs.Has(FeatMCPApps))
}

func TestFeatureSet_HasAll(t *testing.T) {
	fs := FeatureSet{
		FeatFileHostInput: true,
		FeatSourcePath:    true,
		FeatSourceMint:    true,
	}

	// Subset — should return true
	require.True(t, fs.HasAll(FeatureSet{
		FeatFileHostInput: true,
		FeatSourcePath:    true,
	}))

	// Full set — should return true
	require.True(t, fs.HasAll(FeatureSet{
		FeatFileHostInput: true,
		FeatSourcePath:    true,
		FeatSourceMint:    true,
	}))

	// Not subset — should return false
	require.False(t, fs.HasAll(FeatureSet{
		FeatFileHostInput: true,
		FeatMCPApps:       true, // not in fs
	}))

	// Empty requirement set — always true
	require.True(t, fs.HasAll(FeatureSet{}))
}

func TestFeatureSet_EmptySet(t *testing.T) {
	fs := FeatureSet{}

	require.False(t, fs.Has(FeatFileHostInput))
	require.True(t, fs.HasAll(FeatureSet{}))
	require.False(t, fs.HasAll(FeatureSet{FeatFileHostInput: true}))
}

func TestFeatureSet_NilSet(t *testing.T) {
	var fs FeatureSet // nil map

	require.False(t, fs.Has(FeatFileHostInput))
	require.True(t, fs.HasAll(FeatureSet{}))
	require.False(t, fs.HasAll(FeatureSet{FeatFileHostInput: true}))
}

// ---------------------------------------------------------------------------
// Profile method tests
// ---------------------------------------------------------------------------

func TestPlatformProfile_Has(t *testing.T) {
	p := ProfileStdioGeneric

	require.True(t, p.Has(FeatSourcePath))
	require.True(t, p.Has(FeatCoLocated))
	require.True(t, p.Has(FeatSinkLocal))
	require.False(t, p.Has(FeatRemoteAccess))
	require.False(t, p.Has(FeatFileHostInput))
}

func TestPlatformProfile_IsTransport(t *testing.T) {
	require.True(t, ProfileStdioGeneric.IsTransport(TransportStdio))
	require.False(t, ProfileStdioGeneric.IsTransport(TransportHTTP))
	require.False(t, ProfileStdioGeneric.IsTransport(TransportOpenAI))

	require.True(t, ProfileOpenAIHTTP.IsTransport(TransportHTTP))
	require.True(t, ProfileOpenAITunnel.IsTransport(TransportOpenAI))
}

func TestPlatformProfile_IsHost(t *testing.T) {
	require.True(t, ProfileStdioGeneric.IsHost(HostGeneric))
	require.False(t, ProfileStdioGeneric.IsHost(HostOpenAI))

	require.True(t, ProfileOpenAITunnel.IsHost(HostChatGPT))
	require.True(t, ProfileOpenAIHTTP.IsHost(HostOpenAI))
	require.True(t, ProfileClaudeHTTP.IsHost(HostClaude))
	require.False(t, ProfileClaudeHTTP.IsHost(HostClaudeDesktop))
	// ProfileStdioMCPApps is the shared declaration for the synthetic
	// HostStdioApps alias target — not Claude Desktop directly.
	require.True(t, ProfileStdioMCPApps.IsHost(HostStdioApps))
	require.False(t, ProfileStdioMCPApps.IsHost(HostClaude))

	require.True(t, ProfileHTTPGeneric.IsHost(HostGeneric))
}

// ---------------------------------------------------------------------------
// detectTransport tests
// ---------------------------------------------------------------------------

func TestTransportMechanismFeaturesDerived(t *testing.T) {
	stdio := transportMechanismFeatures(TransportStdio)
	require.True(t, stdio[FeatSourcePath])
	require.True(t, stdio[FeatSinkLocal])
	require.True(t, stdio[FeatSinkDrop])
	require.True(t, stdio[FeatCoLocated])
	require.False(t, stdio[FeatSourceMint])

	http := transportMechanismFeatures(TransportHTTP)
	require.True(t, http[FeatSourceMint])
	require.True(t, http[FeatSinkLocal])
	require.True(t, http[FeatSinkDrop])
	require.True(t, http[FeatRemoteAccess])
	require.False(t, http[FeatSourceURL])
	require.False(t, http[FeatSourceData])

	openai := transportMechanismFeatures(TransportOpenAI)
	require.True(t, openai[FeatSourceURL])
	require.True(t, openai[FeatSourceData])
	require.True(t, openai[FeatSinkLocal])
	require.False(t, openai[FeatSourceMint])
	require.False(t, openai[FeatRemoteAccess])
}

// TestNewProfileMechanismCannotBeOverridden verifies that a profile's declared
// capability features can never flip a transport-mechanism feature: an HTTP
// host that declares e.g. FeatMCPApps still resolves to mint (not url/data),
// so a host supporting the data/url tools cannot corrupt upload_file's source.

func TestNewProfileMechanismCannotBeOverridden(t *testing.T) {
	// A hypothetical HTTP host that renders MCP Apps (like Grok) must still
	// present the HTTP mechanism: mint, never url/data, despite any capability.
	p := newProfile(FeatureSet{FeatMCPApps: true}, HostGrok, TransportHTTP, AuthOAuth, true)
	require.True(t, p.Has(FeatSourceMint))
	require.False(t, p.Has(FeatSourceURL))
	require.False(t, p.Has(FeatSourceData))
	require.True(t, p.Has(FeatMCPApps))
	require.True(t, p.Has(FeatRemoteAccess))
}
