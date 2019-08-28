package alea

import (
    "testing"
    "math"
    "fmt"
)

var
	
	gT *Generator = New()

func TestRandom (t *testing.T) {
	const n = 100
	for i := 0; i < n; i++ {
		fmt.Println(gT.Random())
	}
	fmt.Println()
}

func TestIntRand (t *testing.T) {
	const (n = 100; min = 1000; max = 10000)
	for i := 0; i < n; i++ {
		fmt.Println(gT.IntRand(min, max))
	}
	fmt.Println()
}

func TestGauss (t *testing.T) {
	
	const (
		n = 10000000
		m = 2.
		sig = 3.
	)
	
	var
		x float64
	
	s := 0.; s2 := 0.
	p := 0;
	for i := 1; i <= n; i++ {
		x = m + sig * gT.GaussRand()
		s = s + x
		s2 = s2 + x * x
		if math.Abs(x - m) < sig {
			p += 1
		}
	}
	s = s / n
	s2 = math.Sqrt(s2 / n - s * s)
	fmt.Println("Mean: ", s)
	fmt.Println("Standard deviation: ", s2)
	fmt.Println("prop. at sigma: ", 100. * float64(p) / n)
	fmt.Println()
}
