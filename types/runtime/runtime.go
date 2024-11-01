// Copyright (c) 2024 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

package runtime

import "time"

// RunTime to hold start time, end time and duration of a plugin and/or plugin manager execution.
type RunTime struct {
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
}
