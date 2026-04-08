package services

import (
	"testing"
	"time"
)

func TestBuildSessionClaims(t *testing.T) {
	now := time.Unix(1_700_000_000, 0).UTC()
	claims := BuildSessionClaims("570639", "session-123", now, 24*time.Hour)

	if got := claims["username"]; got != "570639" {
		t.Fatalf("username = %#v, want %q", got, "570639")
	}
	if got := claims["session_id"]; got != "session-123" {
		t.Fatalf("session_id = %#v, want %q", got, "session-123")
	}
}

func TestSessionPolicyDecisions(t *testing.T) {
	now := time.Unix(1_700_000_000, 0).UTC()
	lastActivity := now.Add(-10 * time.Minute)
	revokedAt := now.Add(5 * time.Minute)

	if !ShouldRefreshSessionActivity(&lastActivity, now, 5*time.Minute) {
		t.Fatal("expected refresh interval to trigger")
	}
	if !IsSessionTimedOut(&lastActivity, now.Add(25*time.Minute), 30*time.Minute) {
		t.Fatal("expected session to timeout when idle too long")
	}
	if !IsSessionRevoked(now, &revokedAt) {
		t.Fatal("expected revoked timestamp after token issue to revoke session")
	}
}
