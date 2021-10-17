package ratelimiter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOptionValidation(t *testing.T) {

	valids := []*RateLimiterOptions{
		{},
		{
			IntervalOpt: IntervalOption{
				Interval: time.Minute,
				Limit:    10,
			},
		},
		{
			PeakLoad: 40,
		},
	}

	for _, validOpt := range valids {
		_, err := MakeRateLimiter(validOpt)
		assert.Nil(t, err, "option is valid")
	}

	invalids := []*RateLimiterOptions{
		nil,
		{
			IntervalOpt: IntervalOption{Interval: time.Hour},
		},
		{
			IntervalOpt: IntervalOption{Limit: 42},
		},
	}

	for _, invalidOpt := range invalids {
		_, err := MakeRateLimiter(invalidOpt)
		assert.NotNil(t, err, "option is invalid")
	}
}
