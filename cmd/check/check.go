// check.go
// - test time functions

package main

import (
	"fmt"

	//ftime "github.com/neurocline/gofix/time"
)

func nzero() int64
func nmul(a, b uint64) uint64
func nmulxy(a, b uint64) (x uint64, y uint64)

func main() {
	z := nzero()
	fmt.Printf("Got z=%d\n", z)

	q := nmul(11998413008389, 0x0000006400000000)
	fmt.Printf("Got q: %d\n", q)

	s, t := nmulxy(11998413008389, 0x0000006400000000)
	fmt.Printf("Got s,t: %d, %d\n", s, t)
}
