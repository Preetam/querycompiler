package querycompiler

import (
	"fmt"
	"testing"
)

func TestTokenize(t *testing.T) {
	checkQuery := func(str string) {
		exp, err := readFromTokens(tokenize(str))
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("\n%s\n%s\n", str, compile(exp))
		compile(exp).readRow("")
	}
	checkQuery(`(select (columns name) (table users) (where (= users.id 1) ) )`)
	checkQuery(`(select (columns (count 1)) (table users) (group (users.name (select (columns 1))) )))`)
	checkQuery(`(select (columns 1))`)
	//checkQuery(`(select (name) users ( (= id 1) ) )`)
	//checkQuery(`(select (id) (select (*) users ((exists 1)) ) ( (= 1 1) ) )`)
	//checkQuery(`(select (columns id (select (columns foo) (table bar)) (table users) (where (= id (select (columns max:id))) ))`)
}

// (true)
// (= 1 1)
// (exists (select (1) foo)
