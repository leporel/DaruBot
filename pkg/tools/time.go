package tools

import "time"

func TimeFromMilliseconds(ms int64) time.Time {
	return time.Unix(0, ms*int64(time.Millisecond))
}

func TimeToMilliseconds(t time.Time) int64 {
	return t.UnixNano() / 1000000
}
