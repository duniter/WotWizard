package avl

import (
	
	A	"util/alea"
		"fmt"
		"testing"

)

const predefined = false

type comp int

func (i comp) Compare (j Comparer) Comp {
	switch {
		case i < j.(comp): {
			return Lt
		}
		case i > j.(comp): {
			return Gt
		}
	}
	return Eq
}

func (i comp) Copy () Copier {
	return i
}

var treeCreate func () (*Tree, []comp, int)

func treeCreate1 () (*Tree, []comp, int) {
	var values = []comp{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}
	tree := New()
	for _, val := range values {
		tree.SearchIns(val)
	}
	return tree, values, 0
}

func treeCreate2 () (*Tree, []comp, int){
	const (nb = 1000; max = 5000)
	values := make([]comp, nb)
	tree := New()
	doublons := 0
	for i := 0; i < nb; i++ {
		values[i] = comp(A.IntRand(0, max))
		//fmt.Println(values[i])
		if _, ok, _ := tree.SearchIns(values[i]); ok {
			doublons++
		}
	}
	return tree, values, doublons
}

func init () {
	if predefined {
		treeCreate = treeCreate1
	} else {
		treeCreate = treeCreate2
	}
}

func TestSearchIns (t *testing.T) {
	treeCreate()
}

func TestNext (t *testing.T) {
	tree, values, doublons := treeCreate()
	n := 0
	v1 := comp(-32000)
	e := tree.Next(nil)
	for e != nil {
		n++
		v2 := e.Val().(comp)
		if v2 <= v1 {
			t.Fail()
		}
		v1 = v2
		e = tree.Next(e)
	}
	if n != tree.NumberOfElems() || n != len(values) - doublons {
		t.Fail()
	}
}

func TestPrevious(t *testing.T) {
	tree, values, doublons := treeCreate()
	n := 0
	v1 := comp(+32000)
	e := tree.Previous(nil)
	for e != nil {
		n++
		v2 := e.Val().(comp)
		if v2 >= v1 {
			t.Fail()
		}
		v1 = v2
		e = tree.Previous(e)
	}
	if n != tree.NumberOfElems() || n != len(values) - doublons {
		t.Fail()
	}
}

func TestSearch(t *testing.T) {
	tree, values, _ := treeCreate()
	if _, ok, _ := tree.Search(values[len(values)/2]); !ok {
		t.Fail()
	}
}

func TestDelete(t *testing.T) {
	tree, values, _ := treeCreate()
	val := values[len(values)/2]
	if !tree.Delete(val) {
		t.Fail()
	}
	if tree.Delete(val) {
		t.Fail()
	}
	if _, ok, _ := tree.Search(val); ok {
		t.Fail()
	}
}

func TestFind (t *testing.T) {
	tree, values, _ := treeCreate()
	if _, ok := tree.Find(len(values) / 2); !ok {
		t.Fail()
	}
}

func TestErase (t *testing.T) {
	tree, values, _ := treeCreate()
	rank := len(values)/2
	val, _ := tree.Find(rank)
	if _, ok, n := tree.Search(val.Val().(comp)); !ok || n != rank {
		t.Fail()
	}
	tree.Erase(rank)
	if _, ok, _ := tree.Search(val.Val().(comp)); ok {
		t.Fail()
	}
}

func doLinearize (v interface{}, p ...interface{}) {
	val := v.(comp)
	a := p[0].([]comp)
	i := p[1].(*int)
	a[*i] = val
	(*i)++
}

func equalTrees (t1, t2 *Tree) bool {
	if t1.NumberOfElems() != t2.NumberOfElems() {
		return false
	}
	i1 := 0; i2 :=0
	a1 := make([]comp, t1.NumberOfElems())
	a2 := make([]comp, t2.NumberOfElems())
	t1.WalkThrough(doLinearize, a1, &i1)
	t2.WalkThrough(doLinearize, a2, &i2)
	for i := 0; i < len(a1); i++ {
		if a1[i] != a2[i] {
			return false
		}
	}
	return true
}

func equalTrees2 (t1, t2 *Tree) bool {
	e1 := t1.Next(nil); e2 := t2.Next(nil)
	for e1 != nil && e2 != nil {
		if e1.Val() != e2.Val() {
			return false
		}
		e1 = t1.Next(e1); e2 = t2.Next(e2)
	}
	return e1 == nil && e2 == nil
}

func TestCopy (t *testing.T) {
	tree, _, _ := treeCreate()
	cop := tree.Copy()
	if !equalTrees2(tree, cop) {t.Fail()}
}

func TestSplitCat (t *testing.T) {
	tree, values, _ := treeCreate()
	cop := tree.Copy()
	if !equalTrees(tree, cop) {
		t.Fail()
	}
	tree2 := tree.Split(len(values) / 2)
	tree.Cat(tree2)
	if !equalTrees(tree, cop) {
		t.Fail()
	}
}

func TestNumberOfElems(t *testing.T) {
	tree, values, doublons := treeCreate()
	if tree.NumberOfElems() != len(values) - doublons {
		t.Fail()
	}
}

func doPrint (i interface{}, l ...interface{}) {
	fmt.Println(i.(comp))
}

func ExampleWalkThrough1_2() {
	tree, _, _  := treeCreate1()
	tree.WalkThrough(doPrint)
	// Output:
	// 1
	// 2
	// 3
	// 4
	// 5
	// 6
	// 7
	// 8
	// 9
	// 10
	// 11
	// 12
	// 13
}

func doStruct(e *Elem, t bool) {
	if t {
		fmt.Print(e.val.(comp), " - ", e.rank, " (")
		if e.left.val != nil {
			fmt.Print(e.left.val.(comp), " - ", e.lTag)
		}
		fmt.Print(" * ")
		if e.right.val != nil {
			fmt.Print(e.right.val.(comp), " - ", e.rTag, ")")
		}
		fmt.Println(")")
		doStruct(e.left, e.lTag)
		doStruct(e.right, e.rTag)
	}
}

func ExampleWalkThrough() {
	tree, _, _ := treeCreate1()
	doStruct(tree.root.left, tree.root.lTag)
	// Output:
	// 8 - 8 (4 - true * 10 - true))
	// 4 - 4 (2 - true * 6 - true))
	// 2 - 2 (1 - true * 3 - true))
	// 1 - 1 ( * 2 - false))
	// 3 - 1 (2 - false * 4 - false))
	// 6 - 2 (5 - true * 7 - true))
	// 5 - 1 (4 - false * 6 - false))
	// 7 - 1 (6 - false * 8 - false))
	// 10 - 2 (9 - true * 12 - true))
	// 9 - 1 (8 - false * 10 - false))
	// 12 - 2 (11 - true * 13 - true))
	// 11 - 1 (10 - false * 12 - false))
	// 13 - 1 (12 - false * )
}
