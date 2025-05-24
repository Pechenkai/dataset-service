package services

import "time"

type fakeClock struct{ now time.Time }

func (f fakeClock) Now() time.Time { return f.now }
