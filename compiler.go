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
}

type filterNode struct {
	operator string
	args     []*planNode
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
			columnsExp := listExp[1]
			for _, columnExp := range columnsExp.(List) {
				columnNode := compile(columnExp)
				if columnNode != nil {
					node.columns = append(node.columns, columnNode)
				}
			}

			tableExp := listExp[2]
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

			whereExps := listExp[3].(List)
			for _, whereExp := range whereExps {
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
	if n.table.symbol != "" {
		fmt.Println(prefix + "scan " + n.table.symbol)
	} else {
		n.table.readRow(prefix + "\t")
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
	fmt.Println(prefix + "emit row")
}
