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

import "fmt"

type Node interface {
	Evaluate(env *Environment) *Row
	string(prefix string) string
}

type Row struct {
	Values map[string]Expression
}

// ConstNode represents a constant.
type ConstNode struct {
	Value Expression
}

func (n *ConstNode) Evaluate(env *Environment) *Row {
	return &Row{Values: map[string]Expression{
		"_": n.Value,
	}}
}

func (n *ConstNode) String() string {
	return n.string("")
}

func (n *ConstNode) string(prefix string) string {
	return prefix + fmt.Sprintf("CONST(%s)", n.Value.ExprToStr())
}

// SymbolNode represents a symbol lookup.
type SymbolNode struct {
	Symbol string
}

func (n *SymbolNode) Evaluate(env *Environment) *Row {
	sym, exists := env.Get(n.Symbol)
	if !exists {
		return nil
	}
	return &Row{Values: map[string]Expression{
		"_": sym,
	}}
}

func (n *SymbolNode) String() string {
	return n.string("")
}

func (n *SymbolNode) string(prefix string) string {
	return prefix + fmt.Sprintf("SYMBOL(%s)", n.Symbol)
}

// TableNode represents a table read.
type TableNode struct {
	TableName string
}

func (n *TableNode) Evaluate(env *Environment) *Row {
	table := env.GetTable(n.TableName)
	if table == nil {
		return nil
	}
	return &table[0]
}

func (n *TableNode) String() string {
	return n.string("")
}

func (n *TableNode) string(prefix string) string {
	return prefix + fmt.Sprintf("TABLE(%s)", n.TableName)
}

// AggregateNode represents an aggregation.
type AggregateNode struct {
	Function  string
	Arguments []Node
}

func (n *AggregateNode) Evaluate(env *Environment) *Row {
	return &Row{Values: map[string]Expression{
		"_": Number(1),
	}}
}

func (n *AggregateNode) String() string {
	return n.string("")
}

func (n *AggregateNode) string(prefix string) string {
	str := prefix + fmt.Sprintf("AGGREGATE(%s)", n.Function)
	for _, arg := range n.Arguments {
		str += "\n" + arg.string(prefix+" arg ")
	}
	return str
}

// ScanNode returns rows.
type ScanNode struct {
	Columns []Node
	Source  Node
	Filters []Filter
}

func (n *ScanNode) Evaluate(env *Environment) *Row {
	r := &Row{
		Values: map[string]Expression{},
	}
	if n.Source != nil {
		r = n.Source.Evaluate(env)

		for key, val := range r.Values {
			env.Set(key, val)
		}
	}

	for _, col := range n.Columns {
		newEnv := NewEnvironment(env)
		columnResult := col.Evaluate(newEnv)
		if columnResult == nil {
			continue
		}
		for key, val := range columnResult.Values {
			r.Values[key] = val
		}
	}

	return r
}

func (n *ScanNode) String() string {
	return n.string("")
}

func (n *ScanNode) string(prefix string) string {
	str := prefix + "SCAN"
	if n.Source != nil {
		str += "\n" + n.Source.string(prefix+"   -> ")
	}
	for _, fil := range n.Filters {
		str += "\n" + "   Filter: " + fil.Operator
		for _, arg := range fil.Arguments {
			str += "\n" + arg.string(prefix+"     - ")
		}
	}
	for _, col := range n.Columns {
		str += "\n" + col.string(prefix+" - ")
	}
	return str
}

type Filter struct {
	Operator  string
	Arguments []Node
}

// JoinNode joins nodes.
type JoinNode struct {
	// Nodes
}

func (n *JoinNode) Evaluate(env *Environment) *Row {
	return nil
}

func (n *JoinNode) String() string {
	return n.string("")
}

func (n *JoinNode) string(prefix string) string {
	return prefix + "JOIN"
}

// GroupNode groups results
type GroupNode struct {
	Columns []Node
	Group   []Node
	Source  Node
}

func (n *GroupNode) Evaluate(env *Environment) *Row {
	return nil
}

func (n *GroupNode) String() string {
	return n.string("")
}

func (n *GroupNode) string(prefix string) string {
	str := prefix + "GROUP"
	str += "\n" + n.Source.string(prefix+"   -> ")
	for _, col := range n.Group {
		str += "\n" + col.string(prefix+" - ")
	}
	for _, col := range n.Columns {
		str += "\n" + col.string(prefix+" - ")
	}
	return str
}

func plan(exp Expression) Node {
	switch exp.(type) {
	case Symbol:
		return &SymbolNode{
			Symbol: string(exp.(Symbol)),
		}
	case String, Number, Bool:
		return &ConstNode{
			Value: exp,
		}
	case List:
		listExp := exp.(List)
		switch listExp[0] {
		case Symbol("count"):
			aggregate := &AggregateNode{
				Function: string(listExp[0].(Symbol)),
			}
			for _, argExp := range listExp[1:] {
				arg := plan(argExp)
				aggregate.Arguments = append(aggregate.Arguments, arg)
			}
			return aggregate
		case Symbol("select"):
			var result Node
			scanNode := &ScanNode{}
			result = scanNode

			var columns []Node

			hasGroup := false
			hasAggregate := false
			for _, nextExp := range listExp[1:] {
				nextExpList := nextExp.(List)
				switch nextExpList[0] {
				case Symbol("columns"):
					for _, columnExp := range nextExpList[1:] {
						columnNode := plan(columnExp)
						if columnNode == nil {
							panic(columnExp.ExprToStr())
						}
						if _, ok := columnNode.(*AggregateNode); ok {
							hasAggregate = true
						}
						columns = append(columns, columnNode)
					}
				case Symbol("table"):
					tableNode := plan(nextExpList[1])
					if _, ok := tableNode.(*SymbolNode); ok {
						tableNode = &TableNode{
							TableName: tableNode.(*SymbolNode).Symbol,
						}
					}
					scanNode.Source = tableNode

				case Symbol("where"):
					for _, whereExp := range nextExpList[1:] {
						whereExpList := whereExp.(List)
						filter := Filter{
							Operator: string(whereExpList[0].(Symbol)),
						}
						for _, arg := range whereExpList[1:] {
							argNode := plan(arg)
							if argNode != nil {
								filter.Arguments = append(filter.Arguments, argNode)
							}
						}
						scanNode.Filters = append(scanNode.Filters, filter)
					}

				case Symbol("group"):
					groupNode := &GroupNode{
						Source:  scanNode,
						Columns: columns,
					}
					for _, columnExp := range nextExpList[1:] {
						columnNode := plan(columnExp)
						if columnNode == nil {
							panic(columnExp.ExprToStr())
						}
						groupNode.Group = append(groupNode.Group, columnNode)
					}

					result = groupNode
				}
				if hasAggregate {
					if !hasGroup {
						hasGroup = true
						groupNode := &GroupNode{
							Source:  scanNode,
							Columns: columns,
						}
						result = groupNode
					}
				}
				if !hasGroup {
					scanNode.Columns = columns
				}
			}
			return result
		}
	}
	return nil
}
