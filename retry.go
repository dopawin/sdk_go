package dopawin

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"net"
	"net/http"
	"time"
)

type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   150 * time.Millisecond,
		MaxDelay:    2 * time.Second,
	}
}

func (c *Client) BetDiceWithRetry(ctx context.Context, req DiceBetRequest, cfg RetryConfig) (*DiceBetResponse, error) {
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 1
	}
	if cfg.BaseDelay <= 0 {
		cfg.BaseDelay = 150 * time.Millisecond
	}
	if cfg.MaxDelay <= 0 {
		cfg.MaxDelay = 2 * time.Second
	}

	var last error
	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		resp, err := c.BetDice(ctx, req)
		if err == nil {
			return resp, nil
		}
		last = err
		if attempt == cfg.MaxAttempts || !retryable(err) {
			break
		}
		delay := backoffDelay(cfg, attempt)
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
	return nil, last
}

func retryable(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusTooManyRequests || apiErr.StatusCode >= 500
	}
	var netErr net.Error
	return errors.As(err, &netErr)
}

func backoffDelay(cfg RetryConfig, attempt int) time.Duration {
	pow := math.Pow(2, float64(attempt-1))
	delay := time.Duration(float64(cfg.BaseDelay) * pow)
	if delay > cfg.MaxDelay {
		delay = cfg.MaxDelay
	}
	jitter := time.Duration(rand.Int63n(int64(delay/3 + 1)))
	return delay + jitter
}
