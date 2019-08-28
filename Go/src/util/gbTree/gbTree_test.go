package gbTree

import (
	
	A	"util/alea"
	L	"util/leftist"
	M	"util/misc"
		"testing"
		"fmt"
		"strings"
		"unicode"
)

const (

	maxInt = 0x7FFFFFFF

)

type (
	
	integer int
	
	integerFac struct {
	}
	
	gesClesI struct {
	}
	
	real float64
	
	realFac struct {
	}
	
	gesClesR struct {
	}
	
	gesClesCase struct{
	}
	
	elem struct {
		end,
		ta int
		page FilePos
	}
	
	elemD struct {
		end int
		s *String
		page FilePos
	}
	
	elemCle struct {
		end int
		cle string
	}
	
	data struct {
	}
	
	dataFac struct {
	}

)

var
	
	gen = A.New()

func (i *integer) Write (w *Writer) {
	w.OutInt32(int32(*i))
}

func (i *integer) Read (r *Reader) {
	*i = integer(r.InInt32())
}

func (integerFac) New (ta int) Data {
	M.Assert(ta == 0 || ta == INS)
	return new(integer)
}

func (*gesClesI) CompP (s1, s2 Data) Comp {
	i1 := s1.(*integer); i2 := s2.(*integer)
	if *i1 < *i2 {
		return Lt
	}
	if *i1 > *i2 {
		return Gt
	}
	return Eq
}

func (*gesClesI) PrefP (k1 Data, k2 *Data) {
}

func (r *real) Write (w *Writer) {
	w.OutFloat64(float64(*r))
}

func (r *real) Read (re *Reader) {
	*r = real(re.InFloat64())
}

func (realFac) New (sz int) Data{
	M.Assert(sz ==0 || sz == RES)
	return new(real)
}

func (*gesClesR) CompP (s1, s2 Data) Comp {
	r1 := s1.(*real); r2 := s2.(*real)
	if *r1 < *r2 {
		return Lt
	}
	if *r1 > *r2 {
		return Gt
	}
	return Eq
}

func (*gesClesR) PrefP (k1 Data, k2 *Data) {
}

func (e1 *elem) Comp (ee2 L.Comparer) L.Comp {
	e2 := ee2.(*elem)
	if e1.end < e2.end {
		return L.First
	}
	if e1.end > e2.end {
		return L.Last
	}
	return L.Equiv
}

func (e1 *elemD) Comp (ee2 L.Comparer) L.Comp {
	e2 := ee2.(*elemD)
	if e1.end < e2.end {
		return L.First
	}
	if e1.end > e2.end {
		return L.Last
	}
	return L.Equiv
}

func (e1 *elemCle) Comp (ee2 L.Comparer) L.Comp {
	e2 := ee2.(*elemCle)
	if e1.end < e2.end {
		return L.First
	}
	if e1.end > e2.end {
		return L.Last
	}
	return L.Equiv
}

func (p *data) Read (r *Reader) {
}

func (p *data) Write (w *Writer) {
}

func (f dataFac) New (ta int) Data {
	return new(data)
}

func lettreOuChiffre (c byte) bool {
	return c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9'
}

func compP (s1, s2 *string) Comp {
	b1 := []byte(strings.ToUpper(*s1)); b2 := []byte(strings.ToUpper(*s2))
	i1 := 0; i2 := 0
	var (c1, c2 byte)
	for {
		for {
			if  i1 == len(b1) {
				c1 = 0
				break
			} else {
				c1 = b1[i1]
			}
			i1++
			if lettreOuChiffre(c1) {
				break
			}
		}
		for {
			if  i2 == len(b2) {
				c2 = 0
				break
			} else {
				c2 = b2[i2]
			}
			i2++
			if lettreOuChiffre(c2) {
				break
			}
		}
		if c1 != c2 || c1 == 0 {
			break
		}
	}
	if c1 < c2 {
		return Lt
	}
	if c1 > c2 {
		return Gt
	}
	if *s1 < *s2 {
		return Lt
	}
	if *s1 > *s2 {
		return Gt
	}
	return Eq
}

func (*gesClesCase) CompP (ss1, ss2 Data) Comp {
	s1 := ss1.(*String); s2 := ss2.(*String)
	return compP(&s1.C, &s2.C);
}

func (man *gesClesCase) PrefP (k1 Data, k2 *Data) {
	*k2 = StringPrefP(k1.(*String), (*k2).(*String), man.CompP)
}

func TestG1 (tt *testing.T) { 
	
	const (
		
		nbElems = 50000
		dureeMin = 100
		dureeMax = 10000
		taMin = 100
		taMax = 1000
		
		nomBase = "BaseL2.dat"
		nbPages = 2000
	
	)
	fmt.Println("TestG1")
	//gen.Randomize(1)
	t := L.New()
	Fac.CloseBase(nomBase)
	ok := Fac.CreateBase(nomBase, 0); M.Assert(ok)
	b := Fac.OpenBase(nomBase, nbPages)
	var f dataFac
	g := b.CreateDataMan(f)
	n := 0
	m := 0
	x := 0
	el := new(L.Elem)
	e := t.First(&el)
	for n < nbElems || e != nil {
		if n % 1000 == 0 {
			fmt.Println(n, "\t", m, "\t", x)
		}
		n++
		if n < nbElems {
			eF := new(elem)
			eF.ta = int(gen.IntRand(taMin, taMax + 1))
			eF.page = g.AllocateSize(eF.ta)
			eF.end = n + int(gen.IntRand(dureeMin, dureeMax + 1))
			t.Insert(eF)
			x += eF.ta
			m++
			e = t.First(&el)
		}
		for e != nil && e.(*elem).end == n {
			eF := e.(*elem)
			x -= eF.ta
			g.EraseData(eF.page)
			t.Erase(el)
			m--
			e = t.First(&el)
		}
	}
	fmt.Println("Size = ", b.End())
	b.CloseBase()
	fmt.Println()
}

func TestG2 (tt *testing.T) { 

	const (
		
		nbElems = 50000
		dureeMin = 100
		dureeMax = 10000
		taMin = 100
		taMax = 1000
		
		nomBase = "BaseLL2.dat"
		nbPages = 2000
		
	)
	
	fmt.Println("TestG2")
	// gen.Randomize(1)
	t := L.New()
	Fac.CloseBase(nomBase)
	_ = Fac.CreateBase(nomBase, 0)
	b := Fac.OpenBase(nomBase, nbPages)
	var f StringFac
	g := b.CreateDataMan(f)
	n := 0
	m := 0
	x := 0
	el := new(L.Elem)
	e := t.First(&el)
	for n < nbElems || e != nil {
		if n % 1000 == 0 {
			fmt.Println(n, "\t", m, "\t", x)
		}
		n++
		if n < nbElems {
			eF := new(elemD)
			eF.s = new(String)
			p := int(gen.IntRand(taMin, taMax + 1))
			by := make([]rune, p)
			for j := 0; j < p; j++ {
				by[j] = rune(gen.IntRand(0x21, 0x7E + 1))
			}
			eF.s.C = string(by)
			eF.page = g.WriteAllocateData(eF.s)
			eF.end = n + int(gen.IntRand(dureeMin, dureeMax + 1))
			t.Insert(eF)
			x += p
			m++
			e = t.First(&el)
		}
		for e != nil && e.(*elemD).end == n {
			eF := e.(*elemD)
			s := g.ReadData(eF.page).(*String)
			M.Assert(s.C == eF.s.C)
			x -= len(eF.s.C)
			g.EraseData(eF.page)
			t.Erase(el)
			m--
			e = t.First(&el)
		}
	}
	fmt.Println("Size = ", b.End())
	b.CloseBase()
	fmt.Println()
}

func TestG3 (tt *testing.T) { 
	fmt.Println("TestG3")
	// gen.Randomize(10)
	const (
		nomBase = "BaseC2.dat"
		nbPages = 2000
		maxCles = 100
		nbCles = 10000
	)
	Fac.CloseBase(nomBase)
	ok := Fac.CreateBase(nomBase, 1); M.Assert(ok, 100)
	b := Fac.OpenBase(nomBase, nbPages); M.Assert(b != nil, 101)
	rI := b.CreateIndex(0)
	b.WritePlace(0, int64(rI))
	g := StringKeyManager()
	var f StringFac
	ind := b.OpenIndex(rI, g, f)
	iw := ind.Writer()
	fmt.Println("nb. of keys = ", nbCles)
	fmt.Println()
	for i := 1; i <= nbCles; i++ {
		if i % 1000 == 0 {
			fmt.Println(i)
		}
		for {
			n := int(gen.IntRand(1, maxCles))
			s := new(String)
			s.C = ""
			for j := 0; j < n; j++ {
				s.C = s.C + string(rune(gen.IntRand(0x21, 0x7E + 1)))
			}
			if !iw.SearchIns(s) {break}
		}
	}
	fmt.Println()
	
	i := int(gen.IntRand(0, nbCles))
	j := int(gen.IntRand(0, nbCles))
	for j == i {
		j = int(gen.IntRand(0, nbCles))
	}
	fmt.Println("nb. of suppressed keys = ", M.Abs(i - j))
	fmt.Println()
	ir := ind.NewReader()
	for m := 0; m < M.Min(i, j); m++ {
		ir.Next()
	}
	for m := 1; m <= M.Abs(i - j); m++ {
		if m % 100 ==  0 {
			fmt.Println(m)
		}
		irc := ir.Clone()
		irc.Next()
		cle := irc.CurrentKey().(*String)
		ok := iw.Erase(cle); M.Assert(ok)
	}
	
	m := 0
	ir.ResetPos()
	ir.Next()
	for ir.PosSet() {
		m++
		//cle := ind.CurrentKey().(*String)
		//fmt.Println(cle.C)
		ir.Next()
	}
	fmt.Println()
	fmt.Println("nb of remaining keys = ", m);
	b.DeleteIndex(rI)
	fmt.Println("Size = ", b.End())
	b.CloseBase()
	fmt.Println()
}

func TestG4 (tt *testing.T) { 
	fmt.Println("TestG4")
	// gen.Randomize(10)
	const (
		nomBase = "BaseI2.dat"
		nbPages = 2000
		nbCles = 10000
	)
	Fac.CloseBase(nomBase)
	ok := Fac.CreateBase(nomBase, 1); M.Assert(ok)
	b := Fac.OpenBase(nomBase, nbPages); M.Assert(b != nil, 101)
	rI := b.CreateIndex(INS)
	b.WritePlace(0, int64(rI))
	var ger gesClesI
	g := MakeKM(&ger)
	var f integerFac
	ind := b.OpenIndex(FilePos(b.ReadPlace(0)), g, f)
	iw := ind.Writer()
	for i := 1; i <= nbCles; i++ {
		if i % 1000 == 0 {
			fmt.Println(i)
		}
		j := new(integer)
		*j = integer(gen.IntRand(0, maxInt))
		for iw.SearchIns(j) {
			*j = integer(gen.IntRand(0, maxInt))
		}
	}
	fmt.Println()
	
	i := int(gen.IntRand(0, nbCles))
	for m := 0; m < i; m++ {
		iw.ResetPos()
		iw.Next()
		cle := iw.CurrentKey()
		ok := iw.Erase(cle); M.Assert(ok)
	}
	
	iw.ResetPos()
	iw.Next()
	for iw.PosSet() {
		i := iw.CurrentKey().(*integer)
		fmt.Println(*i)
		iw.Next()
	}
	fmt.Println()
	fmt.Println("Size = ", b.End())
	b.DeleteIndex(rI)
	fmt.Println("Size = ", b.End())
	b.CloseBase()
}

func TestG5 (tt *testing.T) { 
	fmt.Println("TestG5")
	// gen.Randomize(2)
	const (
		nbElems = 50000
		dureeMin = 1
		dureeMax = 1000
		taMin = 1
		taMax = 100
		nomBase = "BaseM2.dat"
		nbPages = 2000
	)
	t := L.New()
	Fac.CloseBase(nomBase)
	ok := Fac.CreateBase(nomBase, 1); M.Assert(ok)
	b := Fac.OpenBase(nomBase, nbPages); M.Assert(b != nil, 101)
	rI := b.CreateIndex(0)
	b.WritePlace(0, int64(rI))
	g := StringKeyManager()
	var ff StringFac
	ind := b.OpenIndex(rI, g, ff)
	M.Assert(ind.IsEmpty())
	iw := ind.Writer()
	n := 0
	el := new(L.Elem)
	e := t.First(&el)
	for n < nbElems || e != nil {
		if n % 1000 == 0 {
			fmt.Println(n)
		}
		n++
		if n < nbElems {
			eF := new(elemCle)
			ta := int(gen.IntRand(taMin, taMax + 1))
			b := make(Bytes, ta)
			for i := 0; i < ta; i++ {
				b[i] = byte(gen.IntRand('A', 'Z' + 1))
			}
			eF.cle = string(b)
			cle := &String{C: eF.cle}
			if !iw.SearchIns(cle) {
				eF.end = n + int(gen.IntRand(dureeMin, dureeMax + 1))
				t.Insert(eF)
				e = t.First(&el)
			}
			M.Assert(!ind.IsEmpty())
		}
		for e != nil && e.(*elemCle).end == n {
			eF := e.(*elemCle)
			cle := &String{C: eF.cle}
			M.Assert(iw.Search(cle))
			ok := iw.Erase(cle); M.Assert(ok)
			M.Assert(!iw.Search(cle))
			t.Erase(el)
			e = t.First(&el)
		}
	}
	M.Assert(ind.IsEmpty())
	fmt.Println("Size = ", b.End())
	b.DeleteIndex(rI)
	fmt.Println("Size = ", b.End())
	b.CloseBase()
	fmt.Println()
}

func TestG6 (tt *testing.T) { 
	fmt.Println("TestG6")
	// gen.Randomize(2)
	const (
		nbElems = 50000
		dureeMin = 100
		dureeMax = 2000
		taMin = 1
		taMax = 100
		nomBase = "BaseN2.dat"
		nbPages = 2000
	)
	t := L.New()
	Fac.CloseBase(nomBase)
	ok := Fac.CreateBase(nomBase, 1); M.Assert(ok)
	b := Fac.OpenBase(nomBase, nbPages); M.Assert(b != nil, 101)
	rI := b.CreateIndex(0)
	b.WritePlace(0, int64(rI))
	g := StringKeyManager()
	var ff StringFac
	ind := b.OpenIndex(FilePos(b.ReadPlace(0)), g, ff)
	iw := ind.Writer()
	n := 0
	el := new(L.Elem)
	e := t.First(&el)
	for n < nbElems {
		if n % 1000 == 0 {
			fmt.Println(n)
		}
		n++
		eF := new(elemCle)
		ta := int(gen.IntRand(taMin, taMax + 1))
		b := make(Bytes, ta)
		for i := 0; i < ta; i++ {
			b[i] = byte(gen.IntRand('a', 'z' + 1))
		}
		eF.cle = string(b)
		cle := &String{C: eF.cle}
		if !iw.SearchIns(cle) {
			eF.end = n + int(gen.IntRand(dureeMin, dureeMax + 1))
			t.Insert(eF)
			e = t.First(&el)
		}
		for e != nil && e.(*elemCle).end == n {
			eF := e.(*elemCle)
			cle := &String{C: eF.cle}
			ok := iw.Erase(cle); M.Assert(ok)
			t.Erase(el)
			e = t.First(&el)
		}
	}
	iw.Next()
	for iw.PosSet() {
		cle := iw.CurrentKey().(*String)
		fmt.Println(cle.C); fmt.Println()
		iw.Next()
	}
	b.DeleteIndex(rI)
	fmt.Println("Size = ", b.End())
	b.CloseBase()
}

func TestG7 (tt *testing.T) { 
	fmt.Println("TestG7")
	// gen.Randomize(10)
	const (
		nomBase = "BaseJ2.dat"
		nbPages = 2000
		nbCles = 10000
	)
	Fac.CloseBase(nomBase)
	ok := Fac.CreateBase(nomBase, 1); M.Assert(ok)
	b := Fac.OpenBase(nomBase, nbPages); M.Assert(b != nil, 101)
	rI := b.CreateIndex(RES)
	b.WritePlace(0, int64(rI))
	var ger gesClesR
	g := MakeKM(&ger)
	var f realFac
	ind := b.OpenIndex(rI, g, f)
	iw := ind.Writer()
	for i := 0; i < nbCles; i++ {
		r := new(real)
		*r = real(gen.Random())
		for iw.SearchIns(r) {
			*r = real(gen.Random())
		}
	}
	
	i := int(gen.IntRand(0, nbCles))
	for m := 0; m < i; m++ {
		iw.ResetPos()
		iw.Next()
		clef := iw.CurrentKey()
		ok := iw.Erase(clef); M.Assert(ok)
	}
	
	iw.ResetPos()
	iw.Next()
	for iw.PosSet() {
		cle := iw.CurrentKey().(*real)
		fmt.Println(float64(*cle))
		iw.Next()
	}
	fmt.Println()
	b.DeleteIndex(rI)
	b.CloseBase()
}

func TestG8 (tt *testing.T) { 
	fmt.Println("TestG8")
	gen.Randomize(2)
	const (
		nbElems = 50000
		dureeMin = 1
		dureeMax = 1000
		taMin = 1
		taMax = 100
		nomBase = "BaseCC2.dat"
		nbPages = 2000
		minChar = 'a'
		nbChar = 26
		pAlpha = 0.8
		chMin = 0x21
		chMax = 0x7E
	)
	t := L.New()
	Fac.CloseBase(nomBase)
	ok := Fac.CreateBase(nomBase, 1); M.Assert(ok)
	b := Fac.OpenBase(nomBase, nbPages); M.Assert(b != nil, 101)
	rI := b.CreateIndex(0)
	b.WritePlace(0, int64(rI))
	var ger gesClesCase
	g := MakeKM(&ger)
	var ff StringFac
	ind := b.OpenIndex(rI, g, ff)
	iw := ind.Writer()
	n := 0;
	el := new(L.Elem)
	e := t.First(&el)
	for n < nbElems {
		if n % 1000 == 0 {
			fmt.Println(n)
		}
		n++
		eF := new(elemCle)
		ta := int(gen.IntRand(taMin, taMax + 1))
		b := make(Bytes, ta)
		for i := 0; i < ta; i++ {
			var c rune
			if gen.Random() < pAlpha {
				c = rune(gen.IntRand(minChar, minChar + nbChar))
				if gen.Random() < 0.5 {
					c = unicode.ToUpper(c)
				}
			} else {
				for {
					c = rune(gen.IntRand(chMin, chMax + 1))
					if !(unicode.IsLetter(c) || unicode.IsDigit(c)) {
						break
					}
				}
			}
			b[i] = byte(c)
		}
		eF.cle = string(b)
		cle := &String{C: eF.cle}
		if !iw.SearchIns(cle) {
			eF.end = n + int(gen.IntRand(dureeMin, dureeMax + 1))
			t.Insert(eF)
			e = t.First(&el)
		}
		for e != nil && e.(*elemCle).end == n {
			cle := &String{C: e.(*elemCle).cle}
			ok := iw.Erase(cle); M.Assert(ok)
			t.Erase(el)
			e = t.First(&el)
		}
	}
	iw.ResetPos()
	iw.Next()
	for iw.PosSet() {
		cle := iw.CurrentKey().(*String)
		fmt.Println(cle.C); fmt.Println()
		iw.Next()
	}
	b.DeleteIndex(rI)
	fmt.Println("Size = ", b.End())
	b.CloseBase()
}
