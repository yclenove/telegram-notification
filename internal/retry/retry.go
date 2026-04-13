package retry

import (
	"context"
	"errors"
	"time"
)

type Config struct {
	MaxAttempts    int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
}

// Do 执行指数退避重试：
// - 第一次失败后等待 InitialBackoff
// - 每轮退避时间翻倍，直到 MaxBackoff
// - 总重试次数由 MaxAttempts 控制
func Do(ctx context.Context, cfg Config, fn func() error) error {
	var lastErr error
	backoff := cfg.InitialBackoff

	for i := 1; i <= cfg.MaxAttempts; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
		}

		if i == cfg.MaxAttempts {
			break
		}
		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
		backoff *= 2
		if backoff > cfg.MaxBackoff {
			backoff = cfg.MaxBackoff
		}
	}
	if lastErr == nil {
		return errors.New("retry failed without underlying error")
	}
	return lastErr
}
