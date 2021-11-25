package internal

import "time"

// WaitForJobsDuration defines how long to wait between checking jobs.
type WaitForJobsDuration struct {
	// DefaultDuration is the default duration to wait for.
	DefaultDuration time.Duration
	// MultipleJobsDuration is the duration to wait for when there are multiple jobs.
	MultipleJobsDuration time.Duration
	// JobThreshold minimum number of jobs that cause MultipleJobsDuration to be used.
	JobThreshold int
}

// NewWaitForJobsDuration creates a new instance of WaitForJobsDuration, using twice the duration for at least 3 pending jobs.
func NewWaitForJobsDuration(defaultDuration time.Duration) *WaitForJobsDuration {
	// wait for twice as much time when there are multiple jobs to wait for
	return &WaitForJobsDuration{
		DefaultDuration:      defaultDuration,
		MultipleJobsDuration: defaultDuration * 2,
		JobThreshold:         3,
	}
}

func (w *WaitForJobsDuration) GetDuration(numPendingJobs int) time.Duration {
	if numPendingJobs >= w.JobThreshold {
		return w.MultipleJobsDuration
	} else {
		return w.DefaultDuration
	}
}
