package retryer

import "context"

type DummyRetryer struct{}

func (d *DummyRetryer) Retry(ctx context.Context, attempt int) bool {
	// dummy retryer allows only one request
	if attempt < 2 {
		return true
	}

	return false
}
