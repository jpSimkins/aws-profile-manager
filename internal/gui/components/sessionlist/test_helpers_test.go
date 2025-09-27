package sessionlist

import (
	"time"

	"aws-profile-manager/internal/awscli"
)

// makeTestSessions returns n synthetic ActiveSessionInfo values for use in tests.
func makeTestSessions(n int) []awscli.ActiveSessionInfo {
	sessions := make([]awscli.ActiveSessionInfo, n)
	for i := range sessions {
		sessions[i] = awscli.ActiveSessionInfo{
			SessionName: "test-session",
			StartURL:    "https://example.awsapps.com/start",
			Region:      "us-east-1",
			ExpiresAt:   time.Now().Add(time.Hour),
			IsExpired:   false,
		}
	}
	return sessions
}
