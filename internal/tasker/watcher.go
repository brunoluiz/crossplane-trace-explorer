package tasker

import (
	"context"
	"time"
)

// Periodic runs a certain cb function given a specific interval. It will
// always call the cb function at setup.
func Periodic(ctx context.Context, interval time.Duration, cb func() error) error {
	if err := cb(); err != nil {
		return err
	}

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
