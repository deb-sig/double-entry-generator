package importer

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
)

type exprTokenType int

const (
	exprTokenEOF exprTokenType = iota
	exprTokenValue
	exprTokenString
	exprTokenOp
	exprTokenLParen
	exprTokenRParen
)

type exprToken struct {
	typ exprTokenType
	val string
}

type exprParser struct {
	tokens []exprToken
	pos    int
	row    Row
	order  ir.Order
}

func evalWhen(expr string, row Row, order ir.Order) (bool, error) {
	tokens, err := tokenizeWhen(expr)
	if err != nil {
		return false, err
	}
	p := &exprParser{tokens: tokens, row: row, order: order}
	ok, err := p.parseOr()
	if err != nil {
		return false, err
	}
	if p.peek().typ != exprTokenEOF {
		return false, fmt.Errorf("unexpected token %q", p.peek().val)
	}
	return ok, nil
}

func tokenizeWhen(expr string) ([]exprToken, error) {
	tokens := []exprToken{}
	for i := 0; i < len(expr); {
		c := expr[i]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			i++
			continue
		}
		if c == '\'' || c == '"' {
			quote := c
			j := i + 1
			var b strings.Builder
			for ; j < len(expr); j++ {
				if expr[j] == '\\' && j+1 < len(expr) {
					j++
					b.WriteByte(expr[j])
					continue
				}
				if expr[j] == quote {
					break
				}
				b.WriteByte(expr[j])
			}
			if j >= len(expr) {
				return nil, fmt.Errorf("unterminated string")
			}
			tokens = append(tokens, exprToken{typ: exprTokenString, val: b.String()})
			i = j + 1
			continue
		}
		if c == '(' {
			tokens = append(tokens, exprToken{typ: exprTokenLParen, val: "("})
			i++
			continue
		}
		if c == ')' {
			tokens = append(tokens, exprToken{typ: exprTokenRParen, val: ")"})
			i++
			continue
		}
		if expr[i] == '<' && (i+1 >= len(expr) || expr[i+1] != '=') {
			j := i + 1
			for ; j < len(expr); j++ {
				if expr[j] == '>' {
					break
				}
				if expr[j] == '\n' || expr[j] == '\r' {
					break
				}
			}
			if j < len(expr) && expr[j] == '>' {
				j++
				j = scanMethodTail(expr, j)
				tokens = append(tokens, exprToken{typ: exprTokenValue, val: expr[i:j]})
				i = j
				continue
			}
		}
		if op, ok := matchOperator(expr[i:]); ok {
			tokens = append(tokens, exprToken{typ: exprTokenOp, val: op})
			i += len(op)
			continue
		}
		if strings.HasPrefix(expr[i:], "raw[") {
			j := i + len("raw[")
			for ; j < len(expr); j++ {
				if expr[j] == ']' {
					break
				}
			}
			if j >= len(expr) {
				return nil, fmt.Errorf("unterminated raw field")
			}
			j++
			j = scanMethodTail(expr, j)
			tokens = append(tokens, exprToken{typ: exprTokenValue, val: expr[i:j]})
			i = j
			continue
		}
		if expr[i] == '[' {
			j := i + 1
			for ; j < len(expr); j++ {
				if expr[j] == ']' {
					break
				}
			}
			if j >= len(expr) {
				return nil, fmt.Errorf("unterminated field")
			}
			j++
			j = scanMethodTail(expr, j)
			tokens = append(tokens, exprToken{typ: exprTokenValue, val: expr[i:j]})
			i = j
			continue
		}
		j := i
		for j < len(expr) {
			if expr[j] == ' ' || expr[j] == '\t' || expr[j] == '\n' || expr[j] == '\r' {
				break
			}
			if expr[j] == '(' {
				if j == i {
					break
				}
				j++
				for j < len(expr) && expr[j] != ')' {
					j++
				}
				if j < len(expr) && expr[j] == ')' {
					j++
				}
				continue
			}
			if expr[j] == ')' {
				break
			}
			if _, ok := matchOperator(expr[j:]); ok {
				break
			}
			j++
		}
		if j == i {
			return nil, fmt.Errorf("invalid token near %q", expr[i:])
		}
		tokens = append(tokens, exprToken{typ: exprTokenValue, val: expr[i:j]})
		i = j
	}
	tokens = append(tokens, exprToken{typ: exprTokenEOF})
	return tokens, nil
}

func scanMethodTail(expr string, i int) int {
	for i < len(expr) && expr[i] == '.' {
		i++
		for i < len(expr) && isIdentByte(expr[i]) {
			i++
		}
		if i >= len(expr) || expr[i] != '(' {
			continue
		}
		quote := byte(0)
		i++
		for i < len(expr) {
			c := expr[i]
			if quote != 0 {
				if c == '\\' && i+1 < len(expr) {
					i += 2
					continue
				}
				i++
				if c == quote {
					quote = 0
				}
				continue
			}
			if c == '\'' || c == '"' {
				quote = c
				i++
				continue
			}
			i++
			if c == ')' {
				break
			}
		}
	}
	return i
}

func matchOperator(s string) (string, bool) {
	for _, op := range []string{"&&", "||", "==", "!=", ">=", "<=", "!~", "~", ">", "<"} {
		if strings.HasPrefix(s, op) {
			return op, true
		}
	}
	return "", false
}

func isIdentByte(c byte) bool {
	return c == '_' || c == '-' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z'
}

func (p *exprParser) parseOr() (bool, error) {
	left, err := p.parseAnd()
	if err != nil {
		return false, err
	}
	for p.peek().typ == exprTokenOp && p.peek().val == "||" {
		p.next()
		right, err := p.parseAnd()
		if err != nil {
			return false, err
		}
		left = left || right
	}
	return left, nil
}

func (p *exprParser) parseAnd() (bool, error) {
	left, err := p.parsePrimary()
	if err != nil {
		return false, err
	}
	for p.peek().typ == exprTokenOp && p.peek().val == "&&" {
		p.next()
		right, err := p.parsePrimary()
		if err != nil {
			return false, err
		}
		left = left && right
	}
	return left, nil
}

func (p *exprParser) parsePrimary() (bool, error) {
	if p.peek().typ == exprTokenLParen {
		p.next()
		ok, err := p.parseOr()
		if err != nil {
			return false, err
		}
		if p.peek().typ != exprTokenRParen {
			return false, fmt.Errorf("missing closing parenthesis")
		}
		p.next()
		return ok, nil
	}
	return p.parseComparison()
}

func (p *exprParser) parseComparison() (bool, error) {
	left := p.next()
	if left.typ != exprTokenValue && left.typ != exprTokenString {
		return false, fmt.Errorf("expected left value")
	}
	op := p.next()
	if op.typ != exprTokenOp || op.val == "&&" || op.val == "||" {
		return false, fmt.Errorf("expected comparison operator")
	}
	right := p.next()
	if right.typ != exprTokenValue && right.typ != exprTokenString {
		return false, fmt.Errorf("expected right value")
	}
	leftVal := p.valueOf(left)
	rightVal := p.valueOf(right)
	return compareValues(leftVal, op.val, rightVal)
}

func (p *exprParser) valueOf(token exprToken) string {
	if token.typ == exprTokenString {
		return token.val
	}
	if strings.HasPrefix(token.val, "raw[") {
		field, suffix := splitRawToken(token.val)
		value := p.row.Raw[field]
		switch suffix {
		case "time":
			if t, err := parseDate(value, ""); err == nil {
				return t.Format("15:04")
			}
		case "date":
			if t, err := parseDate(value, ""); err == nil {
				return t.Format("2006-01-02")
			}
		case "timestamp":
			if t, err := parseDate(value, ""); err == nil {
				return strconv.FormatInt(t.Unix(), 10)
			}
		}
		return value
	}
	if strings.HasPrefix(token.val, "[") {
		return evalColumnString(token.val, p.row, p.order)
	}
	if strings.HasPrefix(token.val, "<") {
		return evalColumnString(token.val, p.row, p.order)
	}
	value := fieldValue(token.val, p.row, p.order)
	if value == "" && !rowHasRawField(p.row, token.val) {
		return token.val
	}
	return value
}

func rowHasRawField(row Row, field string) bool {
	_, ok := row.Raw[field]
	return ok
}

func splitRawToken(value string) (string, string) {
	if !strings.HasPrefix(value, "raw[") {
		return value, ""
	}
	end := strings.Index(value, "]")
	if end < 0 {
		return value, ""
	}
	field := value[len("raw["):end]
	suffix := ""
	if rest := value[end+1:]; strings.HasPrefix(rest, ".") {
		suffix = strings.TrimPrefix(rest, ".")
	}
	return field, suffix
}

func compareValues(left, op, right string) (bool, error) {
	switch op {
	case "==":
		return left == right, nil
	case "!=":
		return left != right, nil
	case "~":
		return strings.Contains(left, right), nil
	case "!~":
		return !strings.Contains(left, right), nil
	case ">", ">=", "<", "<=":
		if l, lOK := parseComparableNumber(left); lOK {
			if r, rOK := parseComparableNumber(right); rOK {
				switch op {
				case ">":
					return l > r, nil
				case ">=":
					return l >= r, nil
				case "<":
					return l < r, nil
				case "<=":
					return l <= r, nil
				}
			}
		}
		switch op {
		case ">":
			return left > right, nil
		case ">=":
			return left >= right, nil
		case "<":
			return left < right, nil
		case "<=":
			return left <= right, nil
		}
	}
	return false, fmt.Errorf("unsupported operator %q", op)
}

func parseComparableNumber(value string) (float64, bool) {
	cleaned := strings.NewReplacer(",", "", "¥", "", "￥", "", "$", "", "CNY", "", "RMB", "").Replace(strings.TrimSpace(value))
	if !regexp.MustCompile(`^[+-]?\d+(\.\d+)?$`).MatchString(cleaned) {
		return 0, false
	}
	n, err := strconv.ParseFloat(cleaned, 64)
	return n, err == nil
}

func (p *exprParser) peek() exprToken {
	if p.pos >= len(p.tokens) {
		return exprToken{typ: exprTokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *exprParser) next() exprToken {
	t := p.peek()
	if p.pos < len(p.tokens) {
		p.pos++
	}
	return t
}
