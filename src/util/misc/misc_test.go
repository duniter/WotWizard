package misc

import (
	
	"util/alea"
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

func TestSets (t *testing.T) {
	s1 := Set(int64(alea.UintRand(0, MaxUint64)))
	s2 := Set(int64(alea.UintRand(0, MaxUint64)))
	sU := Union(s1, s2)
	sI := Inter(s1, s2)
	sS := SymDiff(s1, s2)
	sD := Diff(sU, sI)
	fmt.Println("s1 =", s1)
	fmt.Println("s2 =", s2)
	fmt.Println("sU =", sU)
	fmt.Println("sI =", sI)
	fmt.Println("sD =", sD)
	fmt.Println("sS =", sS)
	Want(sS == sD, t)
	Want(Union(s1, s1) == s1, t)
	Want(Inter(s2, s2) == s2, t)
	Want(Union(s1, FullSet()) == FullSet(), t)
	Want(Inter(s2, EmptySet()) == EmptySet(), t)
	Want(Union(s1, Inter(s1, s2)) == s1, t)
	Want(Inter(s1, Union(s1, s2)) == s1, t)
	Want(Inter(s1, s2) == Diff(s1, Diff(s1, s2)), t)
}
