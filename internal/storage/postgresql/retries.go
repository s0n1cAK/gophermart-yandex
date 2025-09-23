package postgresql

import (
	"context"
	"time"

	"github.com/avast/retry-go/v4"
)

func retryWrapper(ctx context.Context, op func() error) error {
	classifier := NewPostgresErrorClassifier()

	return retry.Do(
		op,
		retry.Context(ctx),
		retry.Attempts(5),
		retry.Delay(500*time.Millisecond),
		retry.DelayType(retry.BackOffDelay),

		retry.RetryIf(func(err error) bool {
			return classifier.Classify(err) != NonRetriable
		}),

		retry.LastErrorOnly(true),
	)
}
