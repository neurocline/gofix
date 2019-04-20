// doc.go

// Package time fixes functionality in the Go System library time package,
// adding nanosecond-level precision to Windows platform time calls.
//
// The API for this package is exactly that of the standard time
// package, with two exceptions:
//
// - wall clock/monotonic clock on Windows is closer to nanosecond precision
// - a Nanoseconds() function is added that directly returns monotonic elapsed nanoseconds
//
// Implementation Notes
//
// The current implementation of the Go time package on Windows uses
// timeGetTime to get elapsed time. This is wrong, and should be changed
// to use QueryPerformanceCounter. The timeGetTime call has a lower ceiling
// of 1ms for precision, and, worse, the timebase is global and can be changed
// by other programs running on the system. In the past, QPC had some flaws
// making it non-monotonic or even changing its frequency when the CPU clocked
// down, but these have been resolved; for modern Windows systems, QPC is the
// preferred high-resolution timer.
//
package time
