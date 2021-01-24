package querycompiler

// Copyright 2021 Preetam Jinka
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import (
	"testing"
)

func TestPlan(t *testing.T) {
	checkQuery := func(str string) {
		exp, err := readFromTokens(tokenize(str))
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("\n%s\n%s\n", str, plan(exp).string(""))
		env := NewEnvironment(nil)
		env.tables["users"] = []Row{
			{
				Values: map[string]Expression{"id": Number(1)},
			},
		}
		//fmt.Println("eval:", plan(exp).Evaluate(env))
	}

	checkQuery(`(select (columns 1))`)
	checkQuery(`(select (columns 1) (table users))`)
	checkQuery(`(select (columns name) (table users) (where (= users.id 1) ) )`)
	checkQuery(`(select (columns (count 1)) (table users) (group users.name (select (columns 1)))))`)
	checkQuery(`(select (columns 1 name) (table users) (group name) (where (= (select (columns 1)) 1)))`)
	checkQuery(`(select (columns name) (table users) (where (= id 1) ) )`)
	checkQuery(`(select (columns id) (table (select (*) (table users))))`)
	checkQuery(`(select (columns id (select (columns foo) (table bar))) (table users) (where (= id (select (columns max:id)))) ))`)
}
