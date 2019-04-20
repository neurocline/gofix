// time.go

package time

import (
	stdtime "time"
	"unsafe"
)

// Alias all the Standard Library types, so we can have as much compatibility
// as possible. If we don't do this, we would be unable to link against other
// packages that hold and pass instances of these types.

type Duration = stdtime.Duration
type Location = stdtime.Location
type Month = stdtime.Month
type ParseError = stdtime.ParseError
type Ticker = stdtime.Ticker
type Time = stdtime.Time
type Timer = stdtime.Timer
type Weekday = stdtime.Weekday

// Now returns the current local time, with nanosecond precision (or the closest
// possible).
func Now() Time {
	return stdtime.Unix(0, unixnanoseconds())
}

// Since returns the time elapsed since t. It is shorthand for time.Now().Sub(t);
// however, we do it directly on underlying nanosecond values.
func Since(t Time) Duration {
	return Duration(UnixNanoseconds() - t.UnixNano())
}

// Nanoseconds returns the current time as nanoseconds since startup
// On the default implementation, this is identical to UnixNanoseconds.
func Nanoseconds() int64 {
	return nanoseconds()
}

// UnixNanoseconds returns the number of nanoseconds since the start of the
// Unix epoch (January 1, 1970 UTC).
func UnixNanoseconds() int64 {
	return unixnanoseconds()
}

// ----------------------------------------------------------------------------------------------
// Internal shenanigans.

// This mirrors the time.Time struct. We always create UTC times, so loc=nil
type fastTime struct {
	wall uint64
	ext  int64
	loc *stdtime.Location
}

const (
	hasMonotonic = 1 << 63
	maxWall      = wallToInternal + (1<<33 - 1) // year 2157
	minWall      = wallToInternal               // year 1885
	nsecMask     = 1<<30 - 1
	nsecShift    = 30
)

const (
	// The unsigned zero year for internal calculations.
	// Must be 1 mod 400, and times before it will not compute correctly,
	// but otherwise can be changed at will.
	absoluteZeroYear = -292277022399

	// The year of the zero Time.
	// Assumed by the unixToInternal computation below.
	internalYear = 1

	// Offsets to convert between internal and absolute or Unix times.
	absoluteToInternal int64 = (absoluteZeroYear - internalYear) * 365.2425 * secondsPerDay
	internalToAbsolute       = -absoluteToInternal

	unixToInternal int64 = (1969*365 + 1969/4 - 1969/100 + 1969/400) * secondsPerDay
	internalToUnix int64 = -unixToInternal

	wallToInternal int64 = (1884*365 + 1884/4 - 1884/100 + 1884/400) * secondsPerDay
	internalToWall int64 = -wallToInternal
)

const (
	secondsPerMinute = 60
	secondsPerHour   = 60 * secondsPerMinute
	secondsPerDay    = 24 * secondsPerHour
	secondsPerWeek   = 7 * secondsPerDay
	daysPer400Years  = 365*400 + 97
	daysPer100Years  = 365*100 + 24
	daysPer4Years    = 365*4 + 1
)

// This only works if the two Time structs are identical. This is likely to be
// the case.
func convert(t fastTime) stdtime.Time {
	return *(*stdtime.Time)(unsafe.Pointer(&t))
}
