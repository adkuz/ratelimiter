package ratelimiter

import "time"

type SimpleFunc func()

type Task struct {
	id uint32

	function SimpleFunc

	complete bool
	end      time.Time
}
