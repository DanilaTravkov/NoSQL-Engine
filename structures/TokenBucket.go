package structures

import (
	"time"
)

type TokenBucket struct {
	tMax          int
	tLeft         int
	lastReset     int64
	resetInterval int64
}

func CreateTokenBucket(size, resetInterval, tokensLeft int, lastReset int64) *TokenBucket {
	token := TokenBucket{size, tokensLeft, lastReset, int64(resetInterval)}
	now := time.Now()
	timestamp := now.Unix()
	if timestamp-token.lastReset > token.resetInterval {
		token.tLeft = token.tMax
		token.lastReset = timestamp
	}
	return &token
}

func (token *TokenBucket) GetTokensLeft() int {
	return token.tLeft
}

func (token *TokenBucket) GetLastReset() int64 {
	return token.lastReset
}
func (token *TokenBucket) Update() bool {
	now := time.Now()
	timestamp := now.Unix()

	if timestamp-token.lastReset > token.resetInterval {
		token.tLeft = token.tMax
		token.lastReset = timestamp
	}
	if token.tLeft <= 0 {
		return false
	}
	token.tLeft -= 1
	return true
}
