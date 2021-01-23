package querycompiler

import (
	"fmt"
	"strings"
)

type planNode struct {
	symbol   string
	constant string
	table    *planNode
	columns  []*planNode
	filters  []*filterNode
	group    *groupNode
}

type filterNode struct {
	operator string
	args     []*planNode
}

type groupNode struct {
	args []*planNode
}

func (n *planNode) string(prefix string) string {
	result := &strings.Builder{}
	if n.symbol != "" {
		fmt.Fprintln(result, prefix+n.symbol)
		return result.String()
	}
	if n.constant != "" {
		fmt.Fprintln(result, prefix+n.constant)
		return result.String()
	}
	if n.table != nil {
		fmt.Fprintln(result, prefix+"Table")
		result.WriteString(n.table.string(prefix + "\t"))
	}
	for _, col := range n.columns {
		fmt.Fprintln(result, prefix+"Column")
		result.WriteString(col.string(prefix + "\t"))
	}
	for _, filter := range n.filters {
		fmt.Fprintln(result, prefix+"Filter")
		result.WriteString(filter.string(prefix + "\t"))
	}
	if n.group != nil {
		fmt.Fprintln(result, prefix+"Group")
		result.WriteString(n.group.string(prefix + "\t"))
	}
	return result.String()
}

func (f *filterNode) string(prefix string) string {
	result := &strings.Builder{}
	fmt.Fprintln(result, prefix+"Operator", f.operator)
	for _, arg := range f.args {
		fmt.Fprintln(result, prefix+"Arg")
		result.WriteString(arg.string(prefix + "\t"))
	}
	return result.String()
}

func (f *filterNode) eval(prefix string) {
	fmt.Println(prefix+"Operator", f.operator)
	for _, arg := range f.args {
		fmt.Println(prefix + "Arg")
		arg.readRow(prefix + "\t")
	}
}

func (g *groupNode) string(prefix string) string {
	result := &strings.Builder{}
	for _, arg := range g.args {
		fmt.Fprintln(result, prefix+"Group arg")
		result.WriteString(arg.string(prefix + "\t"))
	}
	return result.String()
}

func (g *groupNode) eval(prefix string) {
	for _, arg := range g.args {
		fmt.Println(prefix + "Group arg")
		arg.readRow(prefix + "\t")
	}
}

func (n *planNode) String() string {
	return n.string("")
}

func compile(exp Expression) *planNode {
	switch exp.(type) {
	case Symbol:
		return &planNode{
			symbol: string(exp.(Symbol)),
		}
	case String, Number, Bool:
		return &planNode{
			constant: exp.ExprToStr(),
		}
	case List:
		listExp := exp.(List)
		if len(listExp) == 0 {
			return nil
		}
		switch listExp[0] {
		case Symbol("select"):
			node := &planNode{}
			for _, nextExp := range listExp[1:] {
				nextExpList := nextExp.(List)
				switch nextExpList[0] {
				case Symbol("columns"):
					for _, columnExp := range nextExpList[1:] {
						columnNode := compile(columnExp)
						if columnNode != nil {
							node.columns = append(node.columns, columnNode)
						}
					}
				case Symbol("table"):
					tableExp := nextExpList[1]
					switch tableExp.(type) {
					case Symbol:
						node.table = &planNode{
							symbol: string(tableExp.(Symbol)),
						}
					case List:
						tableNode := compile(tableExp)
						if tableNode != nil {
							node.table = tableNode
						}
					}
				case Symbol("where"):
					for _, whereExp := range nextExpList[1:] {
						whereExpList := whereExp.(List)
						filter := &filterNode{
							operator: string(whereExpList[0].(Symbol)),
						}
						for _, arg := range whereExpList[1:] {
							argNode := compile(arg)
							if argNode != nil {
								filter.args = append(filter.args, argNode)
							}
						}
						node.filters = append(node.filters, filter)
					}

				case Symbol("group"):
					for _, groupExp := range nextExpList[1:] {
						groupExpList := groupExp.(List)
						group := &groupNode{}
						for _, arg := range groupExpList {
							argNode := compile(arg)
							if argNode != nil {
								group.args = append(group.args, argNode)
							}
						}
						node.group = group
					}
				}
			}
			return node
		}
	}
	return nil
}

func (n *planNode) readRow(prefix string) {
	if n.constant != "" {
		fmt.Println(prefix+"constant", n.constant)
		return
	}
	if n.symbol != "" {
		fmt.Println(prefix+"symbol", n.symbol)
		return
	}
	fmt.Println(prefix + "begin query")
	if n.table != nil {
		if n.table.symbol != "" {
			fmt.Println(prefix + "cursor read on table " + n.table.symbol)
		} else {
			n.table.readRow(prefix + "\t")
		}
	}
	for _, filter := range n.filters {
		filter.eval(prefix + "  filter: ")
	}
	for _, col := range n.columns {
		if col.constant != "" {
			fmt.Println(prefix + "constant column " + col.constant)
			continue
		}
		if col.symbol != "" {
			fmt.Println(prefix + "read column " + col.symbol)
			continue
		}
		col.readRow("column:" + prefix + " \t")
	}
	if n.group != nil {
		n.group.eval(prefix + "  group: ")
	}
	fmt.Println(prefix + "emit row")
}
