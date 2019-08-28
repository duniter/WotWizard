package basic

import (
	
	M	"util/misc"
		"testing"
		"fmt"

)

func TestToDown (t *testing.T) {
	s := "He! I'm here..."
	fmt.Println(s)
	ss := strip(s)
	fmt.Println(ss)
	M.Want(ss == "HeImhere", t)
	ss = ToDown(s)
	fmt.Println(ss)
	M.Want(ss == "HEIMHERE", t)
}

func TestCompP (t *testing.T) {
	M.Want(CompP("", "") == Eq && CompP(":;u]", "B,*") == Gt && CompP("#a", "%A") == Gt && CompP("#A", "%a") == Lt && CompP("#a", "%a") == Lt && CompP("#a", "a") == Gt, t)
}

func TestPrefix (t *testing.T) {
	fmt.Println(Prefix(ToDown(""), ToDown("")), Prefix(ToDown(":;u]"), ToDown("B,*")), Prefix(ToDown("?a,B"), ToDown("!A...b...c")))
	M.Want(Prefix(ToDown(""), ToDown("")) && !Prefix(ToDown(":;u]"), ToDown("B,*")) && Prefix(ToDown("?a,B"), ToDown("!A...b...c")), t)
}

