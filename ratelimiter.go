package ratelimiter

import (
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
)

type RateLimiter struct {
	options RateLimiterOptions

	channel chan SimpleFunc

	idSerial        uint32
	performCount    uint32
	intervalManager IntervalManager
}

func MakeRateLimiter(opts *RateLimiterOptions) (*RateLimiter, error) {

	if opts == nil {
		return nil, ErrNilOptions
	}

	if err := opts.Validate(); err != nil {
		return nil, errors.Wrap(err, "option error")
	}

	chanSize := DefaultChanSize
	if opts.ChannelSize > 0 {
		chanSize = opts.ChannelSize
	}

	rl := &RateLimiter{
		options: *opts,

		channel: make(chan SimpleFunc, chanSize),

		intervalManager: MakeIntervalManager(opts.IntervalOpt),
	}

	go rl.loop()

	return rl, nil
}

func (r *RateLimiter) Perform(function SimpleFunc) {
	r.channel <- function
}

func (r *RateLimiter) Empty() bool {
	return len(r.channel)+int(atomic.LoadUint32(&r.performCount)) == 0
}

func (r *RateLimiter) getID() uint32 {
	return atomic.AddUint32(&r.idSerial, 1)
}

func (r *RateLimiter) peakCheck() bool {
	return (r.options.PeakLoad == 0) || r.performCount < r.options.PeakLoad
}

func (r *RateLimiter) run(task *Task) {

	defer func() {
		task.complete = true
		task.end = time.Now()

		atomic.AddUint32(&r.performCount, ^uint32(0)) // decrement
	}()

	task.function()
}

func (r *RateLimiter) loop() {

	for {
		for !(r.peakCheck() && r.intervalManager.Check()) {
			time.Sleep(100 * time.Microsecond)
		}

		f := <-r.channel

		r.performCount++

		task := &Task{
			id:       r.getID(),
			function: f,
		}

		r.intervalManager.Add(task.id, task)

		go r.run(task)
	}
}
