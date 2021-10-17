package ratelimiter

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type TestCase struct {
	Generate func() <-chan int
	Delay    func(index int) time.Duration
}

type FuncTimeInterval struct {
	id int

	start time.Time
	end   time.Time
}

func testPositivePeak(t *testing.T, testCase *TestCase, opts *RateLimiterOptions) {

	if opts.PeakLoad == 0 {
		assert.FailNow(t, "testPositivePeak: expected positive PeakLoad, got 0")
	}

	limiter, err := MakeRateLimiter(opts)
	assert.Nil(t, err, "options are not valid")

	var counter uint32
	var mx sync.RWMutex

	perform := func(id int, duration time.Duration) {
		limiter.Perform(func() {
			mx.Lock()
			counter++
			mx.Unlock()

			mx.RLock()
			assert.LessOrEqualf(t, counter, opts.PeakLoad,
				"(%6d) expected %d, got %d", id, opts.PeakLoad, counter)
			mx.RUnlock()

			time.Sleep(duration)

			mx.Lock()
			counter--
			mx.Unlock()
		})
	}

	for i := range testCase.Generate() {
		perform(i, testCase.Delay(i))
	}

	for !limiter.Empty() {
		time.Sleep(10 * time.Microsecond)

		mx.RLock()
		assert.LessOrEqualf(t, counter, opts.PeakLoad,
			"(main) expected %d, got %d", opts.PeakLoad, counter)
		mx.RUnlock()
	}
}

func testIntervalLimit(t *testing.T, testCase *TestCase, opts *RateLimiterOptions) {

	if opts.IntervalOpt.Interval == 0 || opts.IntervalOpt.Limit == 0 {
		assert.FailNow(t, "testIntervalLimit: expected positive Interval and Limit")
	}

	limiter, err := MakeRateLimiter(opts)
	assert.Nil(t, err, "options are not valid")

	intervalChan := make(chan FuncTimeInterval, opts.IntervalOpt.Limit)
	intervals := make([]FuncTimeInterval, 0)

	perform := func(id int, duration time.Duration) {
		limiter.GetChannel() <- func() {
			start := time.Now()

			time.Sleep(duration)

			intervalChan <- FuncTimeInterval{
				id:    id,
				start: start,
				end:   time.Now(),
			}
		}
	}

	go func() {
		for interval := range intervalChan {
			intervals = append(intervals, interval)
		}
	}()

	for i := range testCase.Generate() {
		perform(i, testCase.Delay(i))
	}

	for !limiter.Empty() {
		time.Sleep(10 * time.Microsecond)
	}
	close(intervalChan)

	// analyze time intervals
	var wg sync.WaitGroup
	for i := range intervals {

		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			start := intervals[id].start.Add(-time.Duration(rand.Intn(2)) * time.Millisecond)
			end := start.Add(opts.IntervalOpt.Interval)

			count := uint32(0)
			for _, interval := range intervals {
				if !(interval.end.Before(start) || interval.start.After(end)) {
					count++
				}
			}

			assert.LessOrEqualf(t, count, opts.IntervalOpt.Limit,
				"[%s - %s] catched %d funcs, max is %d",
				start.String(), end.String(),
				count, opts.IntervalOpt.Limit,
			)
		}(i)
	}
	wg.Wait()
}

func TestPeakLoad(t *testing.T) {

	options := []*RateLimiterOptions{
		{PeakLoad: 1},
		{PeakLoad: 10},
		{PeakLoad: 100},
		{
			PeakLoad: 20,
			IntervalOpt: IntervalOption{
				Interval: time.Minute,
				Limit:    100,
			},
		},
		{
			PeakLoad: 5,
			IntervalOpt: IntervalOption{
				Interval: 10 * time.Second,
				Limit:    90,
			},
		},
	}

	cases := []*TestCase{
		{
			Generate: func() <-chan int {
				ch := make(chan int)

				go func(ch chan<- int) {
					for i := 0; i < 100; i++ {
						ch <- i
					}
					close(ch)
				}(ch)

				return ch
			},

			Delay: func(i int) time.Duration {
				return time.Millisecond * time.Duration(rand.Intn(100))
			},
		},
	}

	for _, option := range options {
		for _, testCase := range cases {
			testPositivePeak(t, testCase, option)
		}
	}
}

func TestRateLimit(t *testing.T) {

	options := []*RateLimiterOptions{
		{
			IntervalOpt: IntervalOption{
				Interval: 10 * time.Second,
				Limit:    10,
			},
		},
		{
			IntervalOpt: IntervalOption{
				Interval: 5 * time.Second,
				Limit:    20,
			},
		},
		{
			PeakLoad: 200,
			IntervalOpt: IntervalOption{
				Interval: 30 * time.Second,
				Limit:    45,
			},
		},
	}

	cases := []*TestCase{
		{
			Generate: func() <-chan int {
				ch := make(chan int)

				go func(ch chan<- int) {
					for i := 0; i < 100; i++ {
						time.Sleep(time.Millisecond * time.Duration(rand.Intn(10)))
						ch <- i
					}
					close(ch)
				}(ch)

				return ch
			},

			Delay: func(i int) time.Duration {
				return time.Millisecond * time.Duration(rand.Intn(200))
			},
		},
	}

	for _, option := range options {
		for _, testCase := range cases {
			testIntervalLimit(t, testCase, option)
		}
	}
}
