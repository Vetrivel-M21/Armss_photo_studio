package ratelimit

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"armss-poster-studio/internal/store/mysql"
)

type Limiter struct {
	limitsRepo  *mysql.UserLimitsRepo
	recordsRepo *mysql.RateRecordsRepo
	logger      *slog.Logger
}

func NewLimiter(limitsRepo *mysql.UserLimitsRepo, recordsRepo *mysql.RateRecordsRepo, logger *slog.Logger) *Limiter {
	return &Limiter{
		limitsRepo:  limitsRepo,
		recordsRepo: recordsRepo,
		logger:      logger,
	}
}

type CheckResult struct {
	Allowed           bool
	CurrentUsage      int
	MaxLimit          int
	ErrorMessage      string
	RetryAfterSeconds int
}

// CheckAndRecord evaluates whether the user has remaining quota within the rolling 1-hour window.
// If allowed, it immediately records the action into api_rate_records.
func (l *Limiter) CheckAndRecord(ctx context.Context, userID, actionType string) (CheckResult, error) {
	limits, err := l.limitsRepo.GetEffectiveLimit(ctx, userID)
	if err != nil {
		l.logger.Error("failed_to_get_effective_limit", "user_id", userID, "error", err)
		// On DB error, fail open so users aren't abruptly blocked by quota service errors
		return CheckResult{Allowed: true}, nil
	}

	if limits.IsUnlimited {
		// Log action for reporting without enforcing limit
		_ = l.recordsRepo.RecordAction(ctx, userID, actionType)
		return CheckResult{Allowed: true}, nil
	}

	maxLimit := limits.MaxAPICallsPerHour
	actionLabel := "AI agent operations"
	if actionType == mysql.ActionRender {
		maxLimit = limits.MaxRendersPerHour
		actionLabel = "image renders"
	}

	used, err := l.recordsRepo.CountInPastHour(ctx, userID, actionType)
	if err != nil {
		l.logger.Error("failed_to_count_rate_records", "user_id", userID, "error", err)
		return CheckResult{Allowed: true}, nil
	}

	if used >= maxLimit {
		oldest, err := l.recordsRepo.OldestInPastHour(ctx, userID, actionType)
		var retrySecs int
		var waitMins int = 1
		if err == nil {
			resetAt := oldest.Add(1 * time.Hour)
			remaining := time.Until(resetAt)
			if remaining > 0 {
				retrySecs = int(remaining.Seconds())
				waitMins = int(math.Ceil(remaining.Minutes()))
			}
		}

		errMsg := fmt.Sprintf("Hourly limit reached (%d of %d %s used). Quota resets in %d minutes.", used, maxLimit, actionLabel, waitMins)
		return CheckResult{
			Allowed:           false,
			CurrentUsage:      used,
			MaxLimit:          maxLimit,
			ErrorMessage:      errMsg,
			RetryAfterSeconds: retrySecs,
		}, nil
	}

	if err := l.recordsRepo.RecordAction(ctx, userID, actionType); err != nil {
		l.logger.Error("failed_to_record_rate_action", "user_id", userID, "error", err)
	}

	return CheckResult{
		Allowed:      true,
		CurrentUsage: used + 1,
		MaxLimit:     maxLimit,
	}, nil
}
