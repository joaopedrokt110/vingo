package vingo

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// -------------------- Expression Evaluator (basit) --------------------
//
// Supports:
// - Comparisons: ==, !=, >, <, >=, <=
// - Logical: and, or (left-to-right, no operator precedence beyond that)
// - Parentheses not supported in this simple evaluator (could be added)
// - Left and right operands can be identifiers (dot notation), quoted strings, numbers, booleans.

var compOpRe = regexp.MustCompile(`\s*(==|!=|>=|<=|>|<)\s*`)

func evalCondition(expr string, data map[string]interface{}) (bool, error) {
	// split by " and " / " or " preserving order
	// implement left-to-right evaluation
	tokens := splitLogical(expr)
	if len(tokens) == 0 {
		// treat empty as false
		return false, nil
	}
	// tokens like: [cond, op, cond, op, cond...], where op is "and"/"or"
	// evaluate first cond
	res, err := evalSimpleCond(strings.TrimSpace(tokens[0]), data)
	if err != nil {
		return false, err
	}
	i := 1
	for i < len(tokens)-0 {
		op := strings.TrimSpace(tokens[i])
		nextExpr := strings.TrimSpace(tokens[i+1])
		nextRes, err := evalSimpleCond(nextExpr, data)
		if err != nil {
			return false, err
		}
		if op == "and" {
			res = res && nextRes
		} else if op == "or" {
			res = res || nextRes
		} else {
			return false, fmt.Errorf("unknown logical operator %s", op)
		}
		i += 2
		if i >= len(tokens) {
			break
		}
	}
	return res, nil
}

func splitLogical(expr string) []string {
	// naive split: find " and " and " or " tokens
	parts := []string{}
	cur := ""
	low := strings.TrimSpace(expr)
	words := strings.Fields(low)
	// rebuild by scanning tokens
	i := 0
	for i < len(words) {
		w := words[i]
		if w == "and" || w == "or" {
			parts = append(parts, strings.TrimSpace(cur))
			parts = append(parts, w)
			cur = ""
		} else {
			if cur == "" {
				cur = w
			} else {
				cur += " " + w
			}
		}
		i++
	}
	if cur != "" {
		parts = append(parts, strings.TrimSpace(cur))
	}
	return parts
}

func evalSimpleCond(cond string, data map[string]interface{}) (bool, error) {
	// If condition contains comparison operator -> split
	if compOpRe.MatchString(cond) {
		// loc := compOpRe.FindStringIndex(cond)
		op := compOpRe.FindStringSubmatch(cond)[1]
		parts := compOpRe.Split(cond, 2)
		if len(parts) != 2 {
			return false, fmt.Errorf("invalid comparison in '%s'", cond)
		}
		left := strings.TrimSpace(parts[0])
		right := strings.TrimSpace(parts[1])
		lv, lok := lookup(data, left)
		if !lok {
			// try literal
			lv = literalFromString(left)
		}
		rv, rok := lookup(data, right)
		if !rok {
			rv = literalFromString(right)
		}
		return compareValues(lv, rv, op)
	}
	// no operator => truthy check of the expression (variable or literal)
	v, ok := lookup(data, cond)
	if ok {
		return condTruthy(v), nil
	}
	// maybe it's literal
	v2 := literalFromString(cond)
	return condTruthy(v2), nil
}

func evalConditionWithValue(condExpr string, value interface{}, data map[string]interface{}) (bool, error) {
	// prepare temporary data where lookups can reference __switch__
	tmp := shallowCopyMap(data)
	tmp["__switch__"] = value

	s := strings.TrimSpace(condExpr)
	if s == "" {
		return false, nil
	}

	// shorthand: a single dot or "value" means direct equality to the switch value
	if s == "." || s == "value" || s == "__switch__" {
		// truthy check of value
		return condTruthy(value), nil
	}

	// support comma-separated cases: e.g. "a, b, 3"
	if strings.Contains(s, ",") {
		parts := strings.Split(s, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			// try literal equality first
			lit := literalFromString(p)
			ok, err := compareValues(value, lit, "==")
			if err == nil && ok {
				return true, nil
			}
			// fallback: try evaluating as an expression (can use __switch__ inside)
			res, err := evalCondition(p, tmp)
			if err == nil && res {
				return true, nil
			}
		}
		return false, nil
	}

	// If the case contains a comparison operator, evaluate it with __switch__ available.
	if compOpRe.MatchString(s) {
		// let evalCondition handle lookups like "__switch__ > 5" or ". > 5" if user writes that
		// but replace single "." with __switch__ in expression for convenience:
		expr := strings.ReplaceAll(s, ".", "__switch__")
		return evalCondition(expr, tmp)
	}

	// No operator & no comma: treat as simple literal or identifier.
	// Try literal equality first (numbers/strings/bool)
	lit := literalFromString(s)
	ok, err := compareValues(value, lit, "==")
	if err == nil && ok {
		return true, nil
	}
	// stringified compare
	if fmt.Sprintf("%v", value) == fmt.Sprintf("%v", lit) {
		return true, nil
	}
	// finally try evaluating the expression with __switch__ available (covers cases where
	// condExpr is something like "__switch__ == 5" or complex lookup)
	// also support shorthand where user used '.' inside expression
	expr := strings.ReplaceAll(s, ".", "__switch__")
	res, err := evalCondition(expr, tmp)
	if err == nil {
		return res, nil
	}
	return false, nil
}

func literalFromString(s string) interface{} {
	s = strings.TrimSpace(s)
	// quoted string
	if (strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"")) || (strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'")) {
		unq, err := strconv.Unquote(s)
		if err == nil {
			return unq
		}
		return s[1 : len(s)-1]
	}
	// bool
	if s == "true" {
		return true
	}
	if s == "false" {
		return false
	}
	// int
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	// float
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	// fallback string
	return s
}

func compareValues(a interface{}, b interface{}, op string) (bool, error) {
	// first try numeric comparison
	af, aIsNum := toFloat(a)
	bf, bIsNum := toFloat(b)
	if aIsNum && bIsNum {
		switch op {
		case "==":
			return af == bf, nil
		case "!=":
			return af != bf, nil
		case ">":
			return af > bf, nil
		case "<":
			return af < bf, nil
		case ">=":
			return af >= bf, nil
		case "<=":
			return af <= bf, nil
		}
	}
	// boolean
	if ab, ok := a.(bool); ok {
		if bb, ok2 := b.(bool); ok2 {
			switch op {
			case "==":
				return ab == bb, nil
			case "!=":
				return ab != bb, nil
			}
		}
	}
	// string compare
	as := fmt.Sprintf("%v", a)
	bs := fmt.Sprintf("%v", b)
	switch op {
	case "==":
		return as == bs, nil
	case "!=":
		return as != bs, nil
	case ">":
		return as > bs, nil
	case "<":
		return as < bs, nil
	case ">=":
		return as >= bs, nil
	case "<=":
		return as <= bs, nil
	}
	return false, fmt.Errorf("unsupported comparison between %T and %T", a, b)
}

func toFloat(v interface{}) (float64, bool) {
	switch t := v.(type) {
	case int:
		return float64(t), true
	case int8:
		return float64(t), true
	case int16:
		return float64(t), true
	case int32:
		return float64(t), true
	case int64:
		return float64(t), true
	case uint:
		return float64(t), true
	case uint8:
		return float64(t), true
	case uint16:
		return float64(t), true
	case uint32:
		return float64(t), true
	case uint64:
		return float64(t), true
	case float32:
		return float64(t), true
	case float64:
		return t, true
	default:
		// try parse from string
		if s := fmt.Sprintf("%v", v); s != "" {
			if f, err := strconv.ParseFloat(s, 64); err == nil {
				return f, true
			}
		}
	}
	return 0, false
}

// -------------------- Truthy --------------------

func condTruthy(v interface{}) bool {
	if v == nil {
		return false
	}
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return t != ""
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(v).Int() != 0
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(v).Uint() != 0
	case float32, float64:
		return reflect.ValueOf(v).Float() != 0
	default:
		// slices/maps: non-empty => true
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Slice, reflect.Array, reflect.Map:
			return rv.Len() > 0
		default:
			return true
		}
	}
}

// -------------------- Helpers / utilities --------------------

// lookup: dot notation support for map/struct
func lookup(data map[string]interface{}, path string) (interface{}, bool) {
	// if path is literal string "..." or number or boolean, don't treat as lookup
	p := strings.TrimSpace(path)
	if p == "" {
		return nil, false
	}
	// quoted string?
	if (strings.HasPrefix(p, "\"") && strings.HasSuffix(p, "\"")) || (strings.HasPrefix(p, "'") && strings.HasSuffix(p, "'")) {
		unq, err := strconv.Unquote(p)
		if err == nil {
			return unq, true
		}
	}
	// numeric literal?
	if i, err := strconv.Atoi(p); err == nil {
		return i, true
	}
	if f, err := strconv.ParseFloat(p, 64); err == nil {
		return f, true
	}
	if p == "true" {
		return true, true
	}
	if p == "false" {
		return false, true
	}

	var cur interface{} = data
	parts := strings.Split(p, ".")
	for _, seg := range parts {
		switch node := cur.(type) {
		case map[string]interface{}:
			v, ok := node[seg]
			if !ok {
				return nil, false
			}
			cur = v
		default:
			rv := reflect.ValueOf(cur)
			switch rv.Kind() {
			case reflect.Map:
				if rv.Type().Key().Kind() == reflect.String {
					mv := rv.MapIndex(reflect.ValueOf(seg))
					if !mv.IsValid() {
						return nil, false
					}
					cur = mv.Interface()
				} else {
					return nil, false
				}
			case reflect.Struct:
				f := rv.FieldByName(seg)
				if f.IsValid() {
					cur = f.Interface()
				} else {
					// try method? (not implemented)
					return nil, false
				}
			default:
				return nil, false
			}
		}
	}
	return cur, true
}

func lookupVal(data map[string]interface{}, path string) interface{} {
	v, _ := lookup(data, path)
	return v
}

// -------------------- compile (tokens -> AST nodes) --------------------

func compileTokens(tokens []*Token) ([]Node, error) {
	nodes := []Node{}
	i := 0
	for i < len(tokens) {
		t := tokens[i]
		switch t.Type {
		case TText:
			nodes = append(nodes, &TextNode{Text: t.Value})
			i++
		case TVar:
			// parse filters from t.Raw maybe in future; currently only default supported.
			filters := []string{}
			// if user wants filters like <{ var | upper }>, varPattern must be extended.
			nodes = append(nodes, &VarNode{Name: t.Value, Default: t.Default, Filters: filters})
			i++
		case TIf:
			ifNode, ni, err := parseIf(tokens, i)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, ifNode)
			i = ni
		case TFor:
			forNode, ni, err := parseFor(tokens, i)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, forNode)
			i = ni
		case TSwitch:
			switchNode, ni, err := parseSwitch(tokens, i)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, switchNode)
			i = ni
		default:
			return nil, fmt.Errorf("unexpected token %v at position %d (raw: %s)", t.Type, i, t.Raw)
		}
	}
	return nodes, nil
}

func parseIf(tokens []*Token, start int) (*IfNode, int, error) {
	// tokens[start] is TIf
	root := &IfNode{}
	branches := []IfBranch{{Expr: tokens[start].Value, Body: []Node{}}}
	elseBody := []Node{}
	currentBody := &branches[0].Body
	depth := 0

	i := start + 1
	for i < len(tokens) {
		t := tokens[i]
		switch t.Type {
		case TIf:
			// nested if: append token as part of body by compiling nested structure
			nested, ni, err := parseIf(tokens, i)
			if err != nil {
				return nil, 0, err
			}
			*currentBody = append(*currentBody, nested)
			i = ni
			continue
		case TEndIf:
			if depth == 0 {
				// finish
				root.Branches = branches
				root.Else = elseBody
				return root, i + 1, nil
			}
			depth--
			*currentBody = append(*currentBody, &TextNode{Text: t.Value})
		case TElseIf:
			if depth == 0 {
				branches = append(branches, IfBranch{Expr: t.Value, Body: []Node{}})
				currentBody = &branches[len(branches)-1].Body
				i++
				continue
			}
			*currentBody = append(*currentBody, &TextNode{Text: t.Value})
		case TElse:
			if depth == 0 {
				elseBody = []Node{}
				currentBody = &elseBody
				i++
				continue
			}
			*currentBody = append(*currentBody, &TextNode{Text: t.Value})
		case TFor:
			fnode, ni, err := parseFor(tokens, i)
			if err != nil {
				return nil, 0, err
			}
			*currentBody = append(*currentBody, fnode)
			i = ni
			continue
		case TSwitch:
			snode, ni, err := parseSwitch(tokens, i)
			if err != nil {
				return nil, 0, err
			}
			*currentBody = append(*currentBody, snode)
			i = ni
			continue
		default:
			// Text or Var
			switch t.Type {
			case TText:
				*currentBody = append(*currentBody, &TextNode{Text: t.Value})
			case TVar:
				*currentBody = append(*currentBody, &VarNode{Name: t.Value, Default: t.Default})
			default:
				return nil, 0, fmt.Errorf("unexpected token inside if: %v", t.Type)
			}
			i++
		}
	}
	return nil, 0, fmt.Errorf("unclosed if starting at token %d", start)
}

func parseFor(tokens []*Token, start int) (*ForNode, int, error) {
	// tokens[start] is TFor with Value like "idx, item:listExpr" or "item:listExpr"
	parts := strings.SplitN(tokens[start].Value, ":", 2)
	if len(parts) != 2 {
		return nil, 0, fmt.Errorf("invalid for tag: %s", tokens[start].Raw)
	}
	left := strings.TrimSpace(parts[0])
	listExpr := strings.TrimSpace(parts[1])

	indexVar := ""
	itemVar := ""
	if strings.Contains(left, ",") {
		p := strings.SplitN(left, ",", 2)
		indexVar = strings.TrimSpace(p[0])
		itemVar = strings.TrimSpace(p[1])
	} else {
		itemVar = left
	}

	node := &ForNode{IndexVar: indexVar, ItemVar: itemVar, ListExpr: listExpr, Body: []Node{}}
	i := start + 1
	depth := 0
	for i < len(tokens) {
		t := tokens[i]
		switch t.Type {
		case TFor:
			// nested for
			nf, ni, err := parseFor(tokens, i)
			if err != nil {
				return nil, 0, err
			}
			node.Body = append(node.Body, nf)
			i = ni
			continue
		case TEndFor:
			if depth == 0 {
				return node, i + 1, nil
			}
			depth--
			node.Body = append(node.Body, &TextNode{Text: t.Value})
		case TIf:
			ifn, ni, err := parseIf(tokens, i)
			if err != nil {
				return nil, 0, err
			}
			node.Body = append(node.Body, ifn)
			i = ni
			continue
		case TSwitch:
			sn, ni, err := parseSwitch(tokens, i)
			if err != nil {
				return nil, 0, err
			}
			node.Body = append(node.Body, sn)
			i = ni
			continue
		default:
			switch t.Type {
			case TText:
				node.Body = append(node.Body, &TextNode{Text: t.Value})
			case TVar:
				node.Body = append(node.Body, &VarNode{Name: t.Value, Default: t.Default})
			default:
				return nil, 0, fmt.Errorf("unexpected token in for: %v", t.Type)
			}
			i++
		}
	}
	return nil, 0, fmt.Errorf("unclosed for starting at token %d", start)
}

func parseSwitch(tokens []*Token, start int) (*SwitchNode, int, error) {
	node := &SwitchNode{Expr: tokens[start].Value, Cases: []SwitchCase{}, Default: []Node{}}
	i := start + 1
	depth := 0 // İç içe switch'leri takip etmek için
	currentCond := ""
	currentBody := []Node{}

	flushCase := func() {
		if currentCond != "" {
			node.Cases = append(node.Cases, SwitchCase{Cond: currentCond, Body: currentBody})
		} else if len(currentBody) > 0 {
			// currentCond boşsa ve body varsa, bu default case'dir
			node.Default = currentBody
		}
		currentCond = ""
		currentBody = []Node{}
	}

	for i < len(tokens) {
		t := tokens[i]
		switch t.Type {
		case TSwitch:
			// İç içe switch
			depth++
			nn, ni, err := parseSwitch(tokens, i)
			if err != nil {
				return nil, 0, err
			}
			currentBody = append(currentBody, nn)
			i = ni
			continue
		case TEndSwitch:
			if depth == 0 {
				// Ana switch bitiyor
				flushCase()
				return node, i + 1, nil
			}
			// İç içe switch bitiyor
			depth--
			currentBody = append(currentBody, &TextNode{Text: t.Value})
			i++
		case TCase:
			if depth == 0 {
				// Ana switch için yeni case
				flushCase()
				currentCond = t.Value
			} else {
				// İç içe switch için normal token
				currentBody = append(currentBody, &TextNode{Text: t.Value})
			}
			i++
		case TDefault:
			if depth == 0 {
				// Ana switch için default case
				flushCase()
				currentCond = "" // Default case için cond boş
			} else {
				// İç içe switch için normal token
				currentBody = append(currentBody, &TextNode{Text: t.Value})
			}
			i++
		case TIf:
			in, ni, err := parseIf(tokens, i)
			if err != nil {
				return nil, 0, err
			}
			currentBody = append(currentBody, in)
			i = ni
			continue
		case TFor:
			fn, ni, err := parseFor(tokens, i)
			if err != nil {
				return nil, 0, err
			}
			currentBody = append(currentBody, fn)
			i = ni
			continue
		default:
			switch t.Type {
			case TText:
				currentBody = append(currentBody, &TextNode{Text: t.Value})
			case TVar:
				currentBody = append(currentBody, &VarNode{Name: t.Value, Default: t.Default})
			default:
				return nil, 0, fmt.Errorf("unexpected token in switch: %v", t.Type)
			}
			i++
		}
	}
	return nil, 0, fmt.Errorf("unclosed switch starting at token %d", start)
}
