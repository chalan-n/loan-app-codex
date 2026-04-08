package services

import (
	"encoding/json"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// BuildSessionClaims centralizes the shape of issued session JWT claims.
func BuildSessionClaims(username, sessionID string, now time.Time, ttl time.Duration) jwt.MapClaims {
	return jwt.MapClaims{
		"username":   username,
		"session_id": sessionID,
		"iat":        now.Unix(),
		"exp":        now.Add(ttl).Unix(),
	}
}

// ParseIssuedAtClaim extracts the issued-at timestamp from JWT claims.
func ParseIssuedAtClaim(claims jwt.MapClaims) (time.Time, bool) {
	switch value := claims["iat"].(type) {
	case float64:
		return time.Unix(int64(value), 0).UTC(), true
	case int64:
		return time.Unix(value, 0).UTC(), true
	case json.Number:
		v, err := value.Int64()
		if err != nil {
			return time.Time{}, false
		}
		return time.Unix(v, 0).UTC(), true
	default:
		return time.Time{}, false
	}
}

// IsSessionRevoked returns true when the token was issued before the user's revoke timestamp.
func IsSessionRevoked(issuedAt time.Time, revokedAt *time.Time) bool {
	return revokedAt != nil && issuedAt.Before(revokedAt.UTC())
}

// IsSessionTimedOut returns true when the last activity is missing or older than the allowed idle timeout.
func IsSessionTimedOut(lastActivity *time.Time, now time.Time, timeout time.Duration) bool {
	return lastActivity == nil || now.Sub(lastActivity.UTC()) > timeout
}

// ShouldRefreshSessionActivity decides whether the last activity timestamp should be updated.
func ShouldRefreshSessionActivity(lastActivity *time.Time, now time.Time, refreshInterval time.Duration) bool {
	return lastActivity != nil && now.Sub(lastActivity.UTC()) >= refreshInterval
}
