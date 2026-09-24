// Package retry provides a small exponential-backoff loop shared by the two
// external HTTP call sites that need it: the Anthropic agent client and the
// OpenAI image provider.
package retry

import (
	"context"
	"math/rand"
	"time"
)

type Config struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

func DefaultConfig() Config {
	return Config{MaxAttempts: 3, BaseDelay: 500 * time.Millisecond, MaxDelay: 4 * time.Second}
}

// Do calls fn up to cfg.MaxAttempts times. fn returns (retryable bool, err error);
// if retryable is false, Do returns immediately without further attempts.
// Backoff is exponential from BaseDelay, capped at MaxDelay, with full jitter.
func Do(ctx context.Context, cfg Config, fn func(attempt int) (retryable bool, err error)) error {
	var lastErr error
	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		retryable, err := fn(attempt)
		if err == nil {
			return nil
		}
		lastErr = err
		if !retryable || attempt == cfg.MaxAttempts {
			return lastErr
		}

		delay := cfg.BaseDelay * time.Duration(1<<uint(attempt-1))
		if delay > cfg.MaxDelay {
			delay = cfg.MaxDelay
		}
		delay = time.Duration(rand.Int63n(int64(delay)))

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
	return lastErr
}
