package utils

import "time"

func TimeFromMilliseconds(ms int64) time.Time {
	return time.Unix(0, ms*int64(time.Millisecond))
}
