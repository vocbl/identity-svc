package domain

import (
	"fmt"
	"time"
)

func validateAccessTokenDuration(duration time.Duration) error {
	min, max := 1*time.Minute, 2*time.Hour
	if duration >= min && duration <= max {
		return nil
	}
	return fmt.Errorf("invalid access token duration: expected between %v and %v, got %v", min, max, duration)
}

func validateRefreshTokenDuration(duration time.Duration) error {
	min, max := 1*time.Hour, 365*24*time.Hour
	if duration >= min && duration <= max {
		return nil
	}
	return fmt.Errorf("invalid refresh token duration: expected between %v and %v, got %v", min, max, duration)
}
