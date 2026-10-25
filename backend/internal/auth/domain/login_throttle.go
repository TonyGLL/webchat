package domain

import (
	"context"
	"fmt"
	"time"
)

// LoginThrottle slows down password guessing per account (not only per IP, which a distributed
// attacker sidesteps). It is a progressive delay, never a permanent lock, so an attacker cannot
// lock a victim out for good.
type LoginThrottle interface {
	// Check returns how long this attempt must wait (zero when it may proceed).
	Check(ctx context.Context, identifier, ip string) (time.Duration, error)
	// RecordFailure counts a failed attempt (wrong password or unknown account).
	RecordFailure(ctx context.Context, identifier, ip string) error
	// RecordSuccess clears the counters of that identifier and IP.
	RecordSuccess(ctx context.Context, identifier, ip string) error
}

// ThrottledError is returned by login while an account (or an identifier+IP pair) is cooling down.
type ThrottledError struct{ RetryAfter time.Duration }

func (e *ThrottledError) Error() string {
	return fmt.Sprintf("too many failed attempts, retry in %d seconds", int(e.RetryAfter.Seconds())+1)
}

const (
	// Failures from the same identifier+IP start slowing down after this many.
	ThrottlePairThreshold = 5
	// Failures for the identifier from anywhere start slowing down after this many, so a
	// distributed attack is slowed while the real owner (other IP) is hardly affected.
	ThrottleAccountThreshold = 20
	ThrottleMaxDelay         = 5 * time.Minute
	ThrottleWindow           = 15 * time.Minute
)

// ThrottleDelay is the wait imposed after `failures` failures when slowing starts at
// `threshold`: 1 s at the threshold-th failure, doubling up to ThrottleMaxDelay.
func ThrottleDelay(failures, threshold int) time.Duration {
	over := failures - threshold
	if over < 0 {
		return 0
	}
	if over > 9 {
		return ThrottleMaxDelay
	}
	d := time.Second << over
	if d > ThrottleMaxDelay {
		return ThrottleMaxDelay
	}
	return d
}
