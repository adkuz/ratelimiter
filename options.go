package ratelimiter

import (
	"errors"
)

var (
	ErrOptionNoInterval  = errors.New("option Interval must be set when there is the Limit")
	ErrOptionNoLimit     = errors.New("option Limit must be set when there is the Interval")
	ErrOptionChannelSize = errors.New("option ChannelSize must be at least 0")
	ErrNilOptions        = errors.New("nil opts")
)

const DefaultChanSize = int32(512)

type RateLimiterOptions struct {
	IntervalOpt IntervalOption
	PeakLoad    uint32

	ChannelSize int32
}

func (o *RateLimiterOptions) Validate() error {

	if o.IntervalOpt.Limit > 0 && o.IntervalOpt.Interval == 0 {
		return ErrOptionNoInterval
	}

	if o.IntervalOpt.Interval > 0 && o.IntervalOpt.Limit == 0 {
		return ErrOptionNoLimit
	}

	if o.ChannelSize < 0 {
		return ErrOptionChannelSize
	}

	return nil
}
