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

type Environment struct {
	outer  *Environment
	values map[string]Expression
	tables map[string][]Row
}

func NewEnvironment(outer *Environment) *Environment {
	return &Environment{
		outer:  outer,
		values: map[string]Expression{},
		tables: map[string][]Row{},
	}
}

func (env *Environment) Get(key string) (Expression, bool) {
	if v, ok := env.values[key]; ok {
		return v, ok
	}
	if env.outer == nil {
		return Nil{}, false
	}
	return env.outer.Get(key)
}

func (env *Environment) Set(key string, value Expression) {
	env.values[key] = value
}

func (env *Environment) SetOuter(key string, value Expression) {
	if _, ok := env.values[key]; ok {
		env.values[key] = value
		return
	}
	if env.outer != nil {
		env.outer.SetOuter(key, value)
	}
}

func (env *Environment) GetTable(key string) []Row {
	if v, ok := env.tables[key]; ok {
		return v
	}
	if env.outer == nil {
		return nil
	}
	return env.outer.GetTable(key)
}
