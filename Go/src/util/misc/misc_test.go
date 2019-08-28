package misc

import (
	
	"testing"
	"fmt"

)

func Test1 (t *testing.T) {
	fmt.Println("MinInt8 = ", MinInt8)
	fmt.Println("MaxInt8 = ", MaxInt8)
	fmt.Println("MinInt16 = ", MinInt16)
	fmt.Println("MaxInt16 = ", MaxInt16)
	fmt.Println("MinInt32 = ", MinInt32)
	fmt.Println("MaxInt32 = ", MaxInt32)
	fmt.Println("MaxInt64 = ", MaxInt64)
	fmt.Println("MinInt64 = ", MinInt64)
	
	fmt.Println("MinUint8 = ", MinUint8)	
	fmt.Println("MaxUint8 = ", MaxUint8)
	fmt.Println("MinUint16 = ", MinUint16)	
	fmt.Println("MaxUint16 = ", MaxUint16)
	fmt.Println("MinUint32 = ", MinUint32)	
	fmt.Println("MaxUint32 = ", MaxUint32)
	fmt.Println("MinUint64 = ", MinUint64)	
	fmt.Println("MaxUint64 = ", MaxUint64)

}

func Test2 (t *testing.T) {
	const name = "Test InstantCreate"
	f, err := InstantCreate(name); Assert(err == nil, 100)
	fmt.Fprint(f, "Hello")
	err = InstantClose(f); Assert(err == nil, 101)
}

func TestOdd (t *testing.T) {
	Want(!Odd(0), t)
	Want(Odd(1), t)
	Want(!Odd(1000), t)
	Want(Odd(1001), t)
}
