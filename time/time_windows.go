// time_windows.go

// +build windows

package time

import (
	"math/big"
	stdtime "time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")
)

var (
	// QueryPerformanceCounter returns the current value of the performance counter,
	// which could be TSC, or HPET, or the ACPI PMI timer
	// https://msdn.microsoft.com/en-us/library/windows/desktop/ms644904(v=vs.85).aspx
	procQueryPerformanceCounter   = kernel32.NewProc("QueryPerformanceCounter")

	// QueryPerformanceFrequency is the number of QPC clocks per second
	// https://msdn.microsoft.com/en-us/library/windows/desktop/ms644905(v=vs.85).aspx
	procQueryPerformanceFrequency = kernel32.NewProc("QueryPerformanceFrequency")

	// The number of nanoseconds per QPC tick
	// TODO: use assembly language to implement a fast integer mutiply with an implicit
	// divide-by-32bit or divide-by-64bit, to avoid the cost of floating-point conversions;
	// use the big package to compute the conversion multiplier.
	nsPerQPC float64
	nsPerQPCv [3]uint32
	nsPerQPCfp uint64

	// Matching pair of QPC and time.Now values, used to coordinate Windows QPC counts
	// with Go time wallclock time/monotonic time.
	startQPC int64
	startTime stdtime.Time

	// The nanosecond offset from QPC time to Unix epoch time
	unixEpochOffsetNs int64
)

//go:noescape
//go:linkname runtimeNano runtime.nanotime
func runtimeNano() int64

//go:noescape
//go:linkname runtimeNow runtime.now
func runtimeNow() (sec int64, nsec int32, mono int64)

func init() {
	// At startup time, get the frequency that QPC runs at. This is guaranteed
	// to be fixed at system boot, so we can cache it.
	var qpf int64
	ret, _, err := procQueryPerformanceFrequency.Call(uintptr(unsafe.Pointer(&qpf)))
	if ret == 0 {
		panic(err.Error())
	}

	nsPerQPC = 1e9 / float64(qpf)

	// Turn this into a 32.64 multiplier, so that we can just do a handful
	// of integer multiplies and adds instead of an int-float-floatdiv-int chain
	var mul big.Int
	mul.Lsh(big.NewInt(1e9), 64)

	var div big.Int
	div.SetInt64(qpf) // but will always be a positive number

	var res big.Int
	res.Div(&mul, &div)

	// Clumsy code to extract out the big-int into our multiplier form
	resBytes := res.Bytes()
	if len(resBytes) > 12 {
		panic("QPF value out of range")
	}
	for len(resBytes) != 12 {
		resBytes = append([]byte{0}, resBytes...)
	}

	for i := 0; i < len(resBytes); i += 4 {
		nsPerQPCv[i/4] = (uint32(resBytes[i+0]) << 24) | (uint32(resBytes[i+1]) << 16) |
					   (uint32(resBytes[i+2]) << 8) | uint32(resBytes[i+3])
	}

	// Make a 32.32 fixed-point number
	nsPerQPCfp = uint64(nsPerQPCv[0]) << 32 | uint64(nsPerQPCv[1])

	// To match this to Go's time package timebase, get both the start qpc value
	// and the start time.Now value at the same time
	// Note: we only panic on the first call to QPC; we assume that if
	// we can call it once successfully, we can always call it successfully.
	// This lets the Go compiler inline the qpc() call below.
	startTime = stdtime.Now()

	ret, _, err = procQueryPerformanceCounter.Call(uintptr(unsafe.Pointer(&startQPC)))
	if ret == 0 {
		panic(err.Error())
	}

	// Compute the Unix epoch offset, the number of nanoseconds to add to qpc/qpf
	// to get a Unix time.
	unixEpochOffsetNs = startTime.UnixNano() - int64(float64(startQPC) * nsPerQPC)
}

func qpc() int64 {
	var qpctime int64

	procQueryPerformanceCounter.Call(uintptr(unsafe.Pointer(&qpctime)))
	return qpctime
}

func qpf() int64 {
	var qpfvalue int64

	procQueryPerformanceFrequency.Call(uintptr(unsafe.Pointer(&qpfvalue)))
	return qpfvalue
}

// nanoseconds returns the current time as nanoseconds since startup
func nanoseconds() int64 {
	return nanoseconds_int()
}

// nanoseconds_float is the simple float-based version. But it's slow, because
// we do an int-float conversion, a float multiply, and then a float-int conversion.
// Also, we start to have problems if the QPC number gets big, because a double
// only has 53 bits of precision (we need to start from zero at program startup time).
func nanoseconds_float() int64 {
	return int64(float64(qpc()) * nsPerQPC)
}

// nanoseconds_int is the much more complicated integer-math version. It has
// the same overflow issue as the floating-point code, but doesn't have a problem
// with large QPC values (not until the number of nanoseconds reaches 2^63, and
// that's not going to happen until you're running for several hundred years straight).
func nanoseconds_int() int64 {
	return qpc_to_int(qpc())
}

// This is the integer-multiply version. It looks daunting, but it's
// a lot faster. Or not! What! Arg.
// base-2^32 multiply of 32.64 (32-bit integer, 64-bit fractional value)
//   DE.00 (qpc)
// *  A.BC (nsPerQPCv)
func qpc_to_int_old(qpcv int64) int64 {

    counter := qpcv
    counter_lo := counter & 0xFFFFFFFF
    counter_hi := counter >> 32
    // c0w0 := uint64(counter_lo) * uint64(nsPerQPCv[2]) -- can safely discard, at most 1/2 bit will affect result
    c0w1 := uint64(counter_lo) * uint64(nsPerQPCv[1])
    c0w2 := uint64(counter_lo) * uint64(nsPerQPCv[0])
    c1w0 := uint64(counter_hi) * uint64(nsPerQPCv[2])
    c1w1 := uint64(counter_hi) * uint64(nsPerQPCv[1])
    c1w2 := uint64(counter_hi) * uint64(nsPerQPCv[0])

    ticks := (c0w1 >> 32) + c0w2 + (c1w0 >> 32) + c1w1 + (c1w2 << 32)
    ticks += (((c0w1 & 0xFFFFFFFF) + (c1w0 & 0xFFFFFFFF) + uint64(0x80000000)) >> 32) // round up lowest bits

    // TBD this should be done in assembly because the compiler can't figure out
    // that it can use the cheaper 32x32=64 multiply instruction.

    return int64(ticks)
}

func nmul(a, b uint64) uint64

func qpc_to_int(qpcv int64) int64 {
	return int64(nmul(uint64(qpcv), nsPerQPCfp))
}

// UnixNanoseconds returns the number of nanoseconds since the start of the
// Unix epoch (January 1, 1970 UTC)
func unixnanoseconds() int64 {
	return nanoseconds() + unixEpochOffsetNs
}
