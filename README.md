# Go Standard Library fixes

This is a set of fixes, in the form of replacement functions, for
defects in the Go Standard Library.

These fixes depend on specific versions of the Go Standard library,
and are thus tracked against the Go version number. These are written
so that they are compatible with the Go Standard library API, and
thus can be gracefully deprecated when these issues are fixed in
the Go Standard library.

There is Godoc documentation for the packages in this module.

## Nanosecond timing on Windows

The Linux and BSD platforms support nanonsecond-level timing, but
the Windows time code uses `timeGetTime`. This is a mistake on the
part of the Go Standard library authors for two reasons

- The minimum resolution of `timeGetTime` is 1ms
- The `timeGetTime` timebase is global

It's the second reason that makes `timeGetTime` useless; any other process
on the machine can change the timebase for `timeGetTime`.

This package updates both wall clock and monotonic time to have nanoscond
resolution on Windows as well as on Linux/BSD (BSD includes macOS). On
non-Windows on Go 1.10 or higher, this is a package alias.

There are two ways to use this, as a complete replacement for `time`, or
as a surgical fix to get nanosecond-level `time.Time` values.

### New functions

`func Nanoseconds() int64` returns elapsed nanoseconds (typically since bootup).

`func UnixNanoseconds() int64` returns nanoseconds since the Unix epoch.

### Type aliases

The `Time` type in `github.com/neurocline/gofix/time` is a type alias to the
Standard library `time.Time` type. This means that `Time` values created by
the `gofix/time` package are runtime-compatible with any `time.Time` values
created by other code.

### Replace package time

The easiest approach is to change import statements; change this

```
import "time"
```

to this

```
import "github.com/neurocline/gofix/time"
```

This is a drop-in replacement. However, every time call gets redirected to
the underlying Standard library `time` package, and these calls might not
be inlined. Furthermore, if all you want is nanosecond timing for performance
measurement, this is slightly slower.

### Augment package time

The other approach is to add a new import statement for this package. This can be
done in one of two ways. The first assumes that you will change all the relevant
calls to `time.Now()` and `time.Since(t Time)` to use the newer more precise calls:

```
import "time"
import nanotime "github.com/neurocline/gofix/time"
```

This can mean less editing of code, but it also means that you are on the hook to
find and change all `time.Now()` and `(time.Since(t Time)` calls and fix them.

The second assumes that you will change all time calls except `time.Now()` and `time.Since(t Time)`

```
import systime "time"
import "github.com/neurocline/gofix/time"
```

Here, you are forced to change any `time` functionality not shimmed by `gofix/time`.
For code using `time.Now` to do simple profiling, this might be the only calls to `time`
that you were making, and so the cost of the shim calls is irrelevant. This approach
is also the easiest to undo, when the Go Standard Library `time` package is finally fixed
(as someday, it must be).
