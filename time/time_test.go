// time_test.go

package time

import (
	"fmt"
	systime "time"

	"testing"
)

var qpc_result int64
func BenchmarkQPC(b *testing.B) {
	var c int64
	for i := 0; i < b.N; i++ {
		c = qpc()
	}
	qpc_result = c
}

func BenchmarkNanosecondsFloat(b *testing.B) {
	var c int64
	for i := 0; i < b.N; i++ {
		c = nanoseconds_float()
	}
	qpc_result = c
}

func BenchmarkNanosecondsInt(b *testing.B) {
	var c int64
	for i := 0; i < b.N; i++ {
		c = nanoseconds_int()
	}
	qpc_result = c
}

func BenchmarkNanotime(b *testing.B) {
	var c int64
	for i := 0; i < b.N; i++ {
		c = runtimeNano()
	}
	qpc_result = c
}

func TestNanoseconds(t *testing.T) {
	var nanos [10]int64

	for i := 0; i < len(nanos); i++ {
		nanos[i] = Nanoseconds()
	}

	// Make sure the times are monotonically increasing.
	for i := 1; i < len(nanos); i++ {
		if nanos[i-1] >= nanos[i] {
			t.Errorf("Nonmonotic time: old=%d cur=%d\n", nanos[i-1], nanos[i])
		}
	}

	// Make sure each time has a distinct value. If we really have
	// a nanosecond-precision clock, each call will have a different value.
	// TBD this test is dodgy.
	for i := 1; i < len(nanos); i++ {
		if nanos[i-1] == nanos[i] {
			t.Errorf("Nonchanging time: t[%d]=%d t[%d]=%d\n", i-1, nanos[i-1], i, nanos[i])
		}
	}

	// Make sure that the delta for each time is appropriately low.
	// TBD this test is dodgy
	for i := 1; i < len(nanos); i++ {
		delta := nanos[i] - nanos[i-1]
		if delta > 1000 {
			t.Errorf("Unexpected high value: %d\n", delta)
		}
	}
}

func TestQPCFloatInt(t *testing.T) {
	var qpcs [10]int64
	fmt.Printf("nsPerQPC = %f\n", nsPerQPC)
	fmt.Printf("nsPerQPCv = %v\n", nsPerQPCv)
	fmt.Printf("nsPerQPCfp = %d\n", nsPerQPCfp)

	for i := 0; i < len(qpcs); i++ {
		qpcs[i] = qpc()
	}

	for i := 0; i < len(qpcs); i++ {
		nano_int := qpc_to_int(qpcs[i])
		nano_float := int64(float64(qpcs[i]) * nsPerQPC)
		if nano_int != nano_float {
			t.Errorf("Delta %d for qpc %d, expected %d == %d\n", nano_int - nano_float, qpcs[i], nano_float, nano_int)
		}
	}
}

//func TestQPC(t *testing.T) {
//	var qpcs [10]int64
//
//	for i := 0; i < len(qpcs); i++ {
//		qpcs[i] = qpc()
//	}
//
//	qpfv := qpf()
//	fmt.Printf("qpf = %d\n", qpfv)
//
//	for i := 0; i < len(qpcs); i++ {
//		if i == 0 {
//			fmt.Printf("%10d\n", qpcs[i])
//		} else {
//			fmt.Printf("%10d - %d\n", qpcs[i], qpcs[i] - qpcs[i-1])
//		}
//	}
//}

func TestUnixNanoseconds(t *testing.T) {
	var nanos [10]int64
	var times [10]systime.Time

	for i := 0; i < len(nanos); i++ {
		nanos[i] = UnixNanoseconds()
		times[i] = systime.Now()
	}

	// The UnixNanoseconds time needs to be within the quantum of
	// systime.Time.UnixNanos. For now, just make sure it's +/- 1 second.
	for i := 0; i < len(nanos); i++ {
		delta := nanos[i] - times[i].UnixNano()
		if delta < -1e9 || delta > +1e9 {
			t.Errorf("Delta %d out of range: nanotime: %d time %d", delta, nanos[i], times[i].UnixNano())
		}
	}
}
