package extStrings

import (
	
	M	"util/misc"
	T	"testing"
		"fmt"

)

func printS (s String) {
	fmt.Println(s.Convert())
}

func Test1 (tst *T.T) {
	dir := Dir()
	const (c1 = "Hello!"; c2 = ", world")
	const (res1 = c1; res2 = "Hello, world!"; res3 = "HeLLo, worLd!")
	fmt.Println(c1)
	fmt.Println(c2)
	s := dir.New()
	s.Set("Yes");
	w := s.NewWriter()
	w.SetPos(0);
	w.WriteString(c1)
	printS(s)
	if s.Convert() != res1 {
		tst.Fail()
	}
	t := dir.New()
	t.Set(c2)
	s = s.Insert(s.Pos(0, "!"), t)
	printS(s)
	if s.Convert() != res2 {
		tst.Fail()
	}
	u := dir.New()
	u.Set("L")
	i := s.Pos(0, "l")
	for i >= 0 {
		t = s.Extract(0, i)
		s = s.Extract(i + 1, M.MaxInt32)
		t = t.Cat(u)
		s = t.Cat(s)
		i = s.Pos(i + 1, "l")
	}
	printS(s)
	if s.Convert() != res3 {
		tst.Fail()
	}
}
