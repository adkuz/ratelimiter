package ratelimiter

import (
	"time"
)

type IntervalOption struct {
	Interval time.Duration
	Limit    uint32
}

type IntervalManager interface {
	Check() bool
	Add(uint32, *Task) uint32
}

type NullIntervalManager struct{}

func (m *NullIntervalManager) Check() bool {
	return true
}

func (m *NullIntervalManager) Add(id uint32, task *Task) uint32 {
	return 0
}

type SimpleIntervalManager struct {
	interval time.Duration
	limit    uint32

	tasks map[uint32]*Task
	count uint32
}

func (m *SimpleIntervalManager) Check() bool {

	endTime := time.Now().Add(-m.interval)
	for id, task := range m.tasks {
		if task.complete && task.end.Before(endTime) {
			delete(m.tasks, id)
			m.count--
		}
	}
	return m.count < m.limit
}

func (m *SimpleIntervalManager) Add(id uint32, task *Task) uint32 {

	m.tasks[id] = task
	m.count++
	return m.count
}

func MakeIntervalManager(opt IntervalOption) IntervalManager {
	if opt.Interval == 0 || opt.Limit == 0 {
		return &NullIntervalManager{}
	}

	return &SimpleIntervalManager{
		interval: opt.Interval,
		limit:    opt.Limit,

		tasks: make(map[uint32]*Task),
	}
}
