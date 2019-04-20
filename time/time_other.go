// time_other.go

// +build !windows

import (
	stdtime "time"
)

// nanoseconds returns the current time as nanoseconds since startup
// On the default implementation, this is identical to UnixNanoseconds
func nanoseconds() int64 {
	return unixnanoseconds()
}

// unixnanoseconds returns the number of nanoseconds since the start of the
// Unix epoch (January 1, 1970 UTC)
func unixnanoseconds() int64 {
	return stdtime.Now().UnixNano()
}
