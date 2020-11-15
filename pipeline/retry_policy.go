package pipeline

import "time"

type Periodic func(interval time.Time, maxAttemps int)
