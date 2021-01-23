package querycompiler

import (
	"testing"
)

func TestTokenize(t *testing.T) {
	checkQuery := func(str string) {
		exp, err := readFromTokens(tokenize(str))
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%s\n%s", str, compile(exp))
		compile(exp).readRow("")
	}
	//checkQuery(`(select (id) users ( (= 1 1) ) )`)
	//checkQuery(`(select (name) users ( (= id 1) ) )`)
	//checkQuery(`(select (id) (select (*) users ((exists 1)) ) ( (= 1 1) ) )`)
	checkQuery(`(select (id name) users ( (= id (select (max:id) users ()) ) ))`)
}

// (true)
// (= 1 1)
// (exists (select (1) foo)
