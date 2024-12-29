package tasker

import (
	"context"
	"time"
)

func Periodic(ctx context.Context, interval time.Duration, cb func() error) error {
	cb()

	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			if err := cb(); err != nil {
				return err
			}
		case <-ctx.Done():
			ticker.Stop()
			return nil
		}
	}
}
