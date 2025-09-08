package vingo

import (
	"fmt"
	"html"
	"reflect"
	"strings"
)

type Node interface {
	Eval(data map[string]interface{}) string
}

type TextNode struct {
	Text string
}

func (n *TextNode) Eval(data map[string]interface{}) string {
	return n.Text
}

type VarNode struct {
	Name    string
	Default string
	Filters []string
}

func containsFilter(filters []string, name string) bool {
	for _, f := range filters {
		if f == name {
			return true
		}
	}
	return false
}

func (n *VarNode) Eval(data map[string]interface{}) string {
	val, ok := lookup(data, n.Name)
	var out string
	if ok {
		out = fmt.Sprintf("%v", val)
	} else if n.Default != "" {
		out = n.Default
	} else {
		out = ""
	}
	// Apply filters in order
	for _, f := range n.Filters {
		out = applyFilter(f, out)
	}

	// Auto-escape unless explicitly marked raw/safe, or global AutoEscape disabled
	if AutoEscape {
		if !containsFilter(n.Filters, "raw") && !containsFilter(n.Filters, "safe") && !containsFilter(n.Filters, "noescape") {
			out = html.EscapeString(out)
		}
	}
	return out
}

type IfNode struct {
	Branches []IfBranch
	Else     []Node
}

type IfBranch struct {
	Expr string
	Body []Node
}

func (n *IfNode) Eval(data map[string]interface{}) string {
	for _, b := range n.Branches {
		ok, err := evalCondition(b.Expr, data)
		if err == nil && ok {
			return evalNodes(b.Body, data)
		}
	}
	// else
	return evalNodes(n.Else, data)
}

type ForNode struct {
	IndexVar string // optional, can be ""
	ItemVar  string
	ListExpr string
	Body     []Node
}

func (n *ForNode) Eval(data map[string]interface{}) string {
	seq, ok := lookup(data, n.ListExpr)
	if !ok {
		return ""
	}
	v := reflect.ValueOf(seq)
	kind := v.Kind()
	if kind != reflect.Slice && kind != reflect.Array {
		return ""
	}
	length := v.Len()
	out := &strings.Builder{}
	for i := 0; i < length; i++ {
		item := v.Index(i).Interface()
		newData := shallowCopyMap(data)
		if n.IndexVar != "" {
			newData[n.IndexVar] = i
		}
		newData[n.ItemVar] = item
		// loop meta
		loopMeta := map[string]interface{}{
			"Index":  i,
			"First":  i == 0,
			"Last":   i == length-1,
			"Length": length,
		}
		newData["loop"] = loopMeta
		out.WriteString(evalNodes(n.Body, newData))
	}
	return out.String()
}

type SwitchNode struct {
	Expr    string
	Cases   []SwitchCase
	Default []Node
}

type SwitchCase struct {
	Cond string
	Body []Node
}

func (n *SwitchNode) Eval(data map[string]interface{}) string {
	val := lookupVal(data, n.Expr)
	// Try to match with case expressions: we evaluate each case as condition:
	for _, c := range n.Cases {
		// if case expression is a simple literal equal to val -> match
		// Alternatively evaluate case as condition using evalCondition, but allow bare literal too.
		ok, err := evalConditionWithValue(c.Cond, val, data)
		if err == nil && ok {
			return evalNodes(c.Body, data)
		}
	}
	// default
	return evalNodes(n.Default, data)
}

// -------------------- Filters --------------------

func applyFilter(name string, input string) string {
	switch name {
	case "upper":
		return strings.ToUpper(input)
	case "lower":
		return strings.ToLower(input)
	case "escape":
		return html.EscapeString(input)
	case "raw":
		// explicit raw: return as-is
		return input
	case "safe":
		// alias for raw
		return input
	default:
		// unknown filter: passthrough
		return input
	}
}
