package retryer

import (
	"context"
	"log"
	"time"
)

type Retryer interface {
	// Retry returns true, if function should be retried after delay, otherwise returns false
	Retry(ctx context.Context, attempt int) bool
}

type retryer struct {
	numRetries int
	//delay is the time to wait before retrying
	delay int
	// filRatio is a percentage of multiplexer capacity, after which it will restrict retries
	fillRatio float64
	// getFillRatio returns current fill ratio of multiplexer capacity
	getFillRatio func() float64
}

const (
	numRetriesDefault = 3
	delayDefault      = 3
	fillRatioDefault  = 0.8
)

func NewRetryer(numRetries int, delay int, fillRatioInt int, getFillRatio func() float64) Retryer {
	if numRetries == 0 {
		numRetries = numRetriesDefault
	}
	if delay == 0 {
		delay = delayDefault
	}

	var fillRatio float64

	if fillRatioInt < 0 || fillRatioInt > 100 {
		log.Printf("Invalid fill ratio. Using default value: %f", fillRatioDefault)
		fillRatio = fillRatioDefault
	} else {
		fillRatio = float64(fillRatioInt) / 100.0
	}

	return &retryer{
		numRetries:   numRetries,
		delay:        delay,
		fillRatio:    fillRatio,
		getFillRatio: getFillRatio,
	}
}

func (r *retryer) Retry(ctx context.Context, attempt int) bool {
	if attempt >= r.numRetries {
		return false
	}

	log.Printf("current ratio is %f\n", r.getFillRatio())
	if r.getFillRatio() < r.fillRatio {

		select {
		case <-ctx.Done():
			return false
		case <-time.After(time.Duration(r.delay) * time.Second):
		}

		return true
	}

	return false
}
