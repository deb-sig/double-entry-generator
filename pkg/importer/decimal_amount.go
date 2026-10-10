package importer

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
)

// CleanAmount mirrors Mirato import_engine::clean_amount: full-width digits,
// parentheses negatives, currency marks, and grouping separators. Scientific
// notation (e/E) is preserved for ParseDecimal.
func CleanAmount(s string) string {
	raw := strings.TrimSpace(s)
	cleaned := normalizeAmountChars(raw)
	negativeByParen := strings.HasPrefix(cleaned, "(") && strings.HasSuffix(cleaned, ")")
	if negativeByParen {
		cleaned = strings.TrimSpace(cleaned[1 : len(cleaned)-1])
	}
	currencyMarks := []string{
		"¥", "￥", "$", "＄", "元", "人民币",
		"CNY", "RMB", "USDT", "USDC", "USD",
		"cny", "rmb", "usdt", "usdc", "usd",
	}
	for {
		changed := false
		for _, mark := range currencyMarks {
			if strings.HasPrefix(cleaned, mark) {
				cleaned = strings.TrimSpace(cleaned[len(mark):])
				changed = true
				break
			}
			if strings.HasSuffix(cleaned, mark) {
				cleaned = strings.TrimSpace(cleaned[:len(cleaned)-len(mark)])
				changed = true
				break
			}
		}
		if !changed {
			break
		}
	}
	var b strings.Builder
	b.Grow(len(cleaned))
	for _, r := range cleaned {
		switch r {
		case ',', '，', '_', '\'', '’', ' ', '\u00a0', '\u202f', '\u2009':
			continue
		default:
			b.WriteRune(r)
		}
	}
	cleaned = b.String()
	if negativeByParen && !strings.HasPrefix(cleaned, "-") {
		cleaned = "-" + cleaned
	}
	return cleaned
}

func normalizeAmountChars(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r >= '０' && r <= '９':
			b.WriteByte(byte('0' + (r - '０')))
		case r == '．' || r == '。':
			b.WriteByte('.')
		case r == '，' || r == '、':
			b.WriteByte(',')
		case r == '＋':
			b.WriteByte('+')
		case r == '－' || r == '−' || r == '—' || r == '–':
			b.WriteByte('-')
		case r == 'Ｅ':
			b.WriteByte('E')
		case r == 'ｅ':
			b.WriteByte('e')
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// ParseAmountDecimal parses a bill/rule amount into an exact decimal.
// prefix (template AmountPrefix) is stripped before cleaning.
// Values that contain arithmetic operators are evaluated with full precedence;
// incomplete/illegal expressions error instead of being truncated by note stripping.
// Trailing non-numeric annotations like "12.00 (备注)" are stripped; notes that
// contain digits or arithmetic operators are not treated as annotations.
func ParseAmountDecimal(value, prefix string) (ir.Decimal, error) {
	value = strings.TrimSpace(value)
	if prefix = strings.TrimSpace(prefix); prefix != "" {
		value = strings.TrimSpace(strings.TrimPrefix(value, prefix))
	}
	normalized := normalizeAmountChars(value)
	core, _ := splitTrailingNonNumericAnnotation(normalized)
	core = strings.TrimSpace(core)
	if isArithmeticAmountExpr(core) {
		// Do not strip parenthetical "notes" — that would turn "1*(2+3" into "1*".
		return evalArithmeticExpression(core)
	}
	cleaned := CleanAmount(value)
	cleaned, _ = splitTrailingNonNumericAnnotation(cleaned)
	cleaned = strings.TrimSpace(cleaned)
	return ir.ParseDecimal(cleaned)
}

// splitTrailingNonNumericAnnotation removes a trailing "(note)" / "（note）"
// only when the note body has no digits and no arithmetic operators.
// Whole-string parentheses negatives like "(1,025.70)" are left untouched.
func splitTrailingNonNumericAnnotation(s string) (string, bool) {
	s = strings.TrimSpace(s)
	for _, pair := range []struct{ open, close string }{
		{"(", ")"},
		{"（", "）"},
	} {
		if !strings.HasSuffix(s, pair.close) {
			continue
		}
		idx := strings.LastIndex(s, pair.open)
		if idx <= 0 {
			// idx==0 => whole value wrapped (paren-negative), not a trailing note.
			continue
		}
		inner := s[idx+len(pair.open) : len(s)-len(pair.close)]
		if isNonNumericAmountNote(inner) {
			return strings.TrimSpace(s[:idx]), true
		}
	}
	return s, false
}

func isNonNumericAmountNote(inner string) bool {
	for _, r := range inner {
		switch {
		case r >= '0' && r <= '9', r >= '０' && r <= '９':
			return false
		case r == '+' || r == '-' || r == '*' || r == '/' || r == '.':
			return false
		case r == '＋' || r == '－' || r == '−' || r == '．' || r == '。':
			return false
		}
	}
	return true
}

// isArithmeticAmountExpr reports whether value should be evaluated as an expression.
// Whole-string parentheses negatives like "(1,025.70)" are NOT arithmetic.
func isArithmeticAmountExpr(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}
	if strings.HasPrefix(trimmed, "(") && strings.HasSuffix(trimmed, ")") {
		inner := strings.TrimSpace(trimmed[1 : len(trimmed)-1])
		if !containsBinaryArithmetic(inner) && !strings.ContainsAny(inner, "()") {
			return false
		}
	}
	return containsBinaryArithmetic(trimmed) || unbalancedOrInnerParens(trimmed)
}

func unbalancedOrInnerParens(value string) bool {
	if !strings.ContainsAny(value, "()") {
		return false
	}
	trimmed := strings.TrimSpace(value)
	if strings.HasPrefix(trimmed, "(") && strings.HasSuffix(trimmed, ")") {
		inner := trimmed[1 : len(trimmed)-1]
		return strings.ContainsAny(inner, "()") || containsBinaryArithmetic(inner)
	}
	// e.g. "1*(2+3" or "(1+2)*3"
	return true
}

// containsBinaryArithmetic reports + - * / used as binary operators (not leading unary).
func containsBinaryArithmetic(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}
	for i := 0; i < len(trimmed); i++ {
		switch trimmed[i] {
		case '*', '/':
			return true
		case '+', '-':
			if i == 0 {
				continue
			}
			prev := trimmed[i-1]
			if prev == 'e' || prev == 'E' {
				continue
			}
			if unicode.IsSpace(rune(prev)) {
				return true
			}
			if (prev >= '0' && prev <= '9') || prev == ')' {
				return true
			}
		}
	}
	return false
}

// setOrderExactMoney stores authoritative decimal money and a legacy float view.
// ExactMoney is never reconstructed from Money.
func setOrderExactMoney(order *ir.Order, amount ir.Decimal) {
	abs := amount.Abs()
	cp := abs
	order.ExactMoney = &cp
	order.Money = abs.Float64Approx()
}

func orderMoneyDecimal(order ir.Order) (ir.Decimal, bool) {
	if order.ExactMoney != nil {
		return *order.ExactMoney, true
	}
	return ir.Decimal{}, false
}

// evalArithmeticExpression evaluates + - * / with standard precedence and
// parentheses using exact decimals. Inexact division, div-by-zero, Inf/NaN,
// and illegal tokens are rejected.
func evalArithmeticExpression(expr string) (ir.Decimal, error) {
	p := &arithParser{s: strings.TrimSpace(expr)}
	if p.s == "" {
		return ir.Decimal{}, fmt.Errorf("empty arithmetic expression")
	}
	v, err := p.parseExpr()
	if err != nil {
		return ir.Decimal{}, err
	}
	p.skipSpace()
	if p.pos < len(p.s) {
		return ir.Decimal{}, fmt.Errorf("unexpected trailing input %q", p.s[p.pos:])
	}
	return v, nil
}

type arithParser struct {
	s   string
	pos int
}

func (p *arithParser) skipSpace() {
	for p.pos < len(p.s) && unicode.IsSpace(rune(p.s[p.pos])) {
		p.pos++
	}
}

func (p *arithParser) parseExpr() (ir.Decimal, error) {
	left, err := p.parseTerm()
	if err != nil {
		return ir.Decimal{}, err
	}
	for {
		p.skipSpace()
		if p.pos >= len(p.s) {
			return left, nil
		}
		op := p.s[p.pos]
		if op != '+' && op != '-' {
			return left, nil
		}
		p.pos++
		right, err := p.parseTerm()
		if err != nil {
			return ir.Decimal{}, err
		}
		if op == '+' {
			left, err = left.Add(right)
		} else {
			left, err = left.Sub(right)
		}
		if err != nil {
			return ir.Decimal{}, err
		}
	}
}

func (p *arithParser) parseTerm() (ir.Decimal, error) {
	left, err := p.parseUnary()
	if err != nil {
		return ir.Decimal{}, err
	}
	for {
		p.skipSpace()
		if p.pos >= len(p.s) {
			return left, nil
		}
		op := p.s[p.pos]
		if op != '*' && op != '/' {
			return left, nil
		}
		p.pos++
		right, err := p.parseUnary()
		if err != nil {
			return ir.Decimal{}, err
		}
		if op == '*' {
			left, err = left.Mul(right)
		} else {
			left, err = left.DivExact(right)
		}
		if err != nil {
			return ir.Decimal{}, err
		}
	}
}

func (p *arithParser) parseUnary() (ir.Decimal, error) {
	p.skipSpace()
	if p.pos < len(p.s) && (p.s[p.pos] == '+' || p.s[p.pos] == '-') {
		sign := p.s[p.pos]
		p.pos++
		v, err := p.parseUnary()
		if err != nil {
			return ir.Decimal{}, err
		}
		if sign == '-' {
			return v.Neg(), nil
		}
		return v, nil
	}
	return p.parsePrimary()
}

func (p *arithParser) parsePrimary() (ir.Decimal, error) {
	p.skipSpace()
	if p.pos >= len(p.s) {
		return ir.Decimal{}, fmt.Errorf("expected number")
	}
	if p.s[p.pos] == '(' {
		p.pos++
		v, err := p.parseExpr()
		if err != nil {
			return ir.Decimal{}, err
		}
		p.skipSpace()
		if p.pos >= len(p.s) || p.s[p.pos] != ')' {
			return ir.Decimal{}, fmt.Errorf("missing closing parenthesis")
		}
		p.pos++
		return v, nil
	}
	start := p.pos
	if p.s[p.pos] == '+' || p.s[p.pos] == '-' {
		p.pos++
	}
	sawDigit := false
	for p.pos < len(p.s) && p.s[p.pos] >= '0' && p.s[p.pos] <= '9' {
		sawDigit = true
		p.pos++
	}
	if p.pos < len(p.s) && p.s[p.pos] == '.' {
		p.pos++
		for p.pos < len(p.s) && p.s[p.pos] >= '0' && p.s[p.pos] <= '9' {
			sawDigit = true
			p.pos++
		}
	}
	if p.pos < len(p.s) && (p.s[p.pos] == 'e' || p.s[p.pos] == 'E') {
		p.pos++
		if p.pos < len(p.s) && (p.s[p.pos] == '+' || p.s[p.pos] == '-') {
			p.pos++
		}
		expDigits := false
		for p.pos < len(p.s) && p.s[p.pos] >= '0' && p.s[p.pos] <= '9' {
			expDigits = true
			p.pos++
		}
		if !expDigits {
			return ir.Decimal{}, fmt.Errorf("invalid scientific literal")
		}
	}
	if !sawDigit {
		return ir.Decimal{}, fmt.Errorf("expected number near %q", p.s[start:])
	}
	return ir.ParseDecimal(p.s[start:p.pos])
}

// looksLikeArithmetic reports whether value is a pure numeric expression that
// should be evaluated (digits, operators, parentheses, whitespace, e/E).
func looksLikeArithmetic(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}
	hasOp := false
	for i := 0; i < len(trimmed); i++ {
		c := trimmed[i]
		switch {
		case c >= '0' && c <= '9', c == '.', c == ' ', c == '\t':
			continue
		case c == 'e', c == 'E':
			continue
		case c == '+' || c == '-' || c == '*' || c == '/' || c == '(' || c == ')':
			hasOp = true
		default:
			return false
		}
	}
	// Lone number is fine for FormatAmountLike callers via Parse; for
	// evalSimpleArithmetic we only rewrite when an operator/paren is present
	// or unary sign wrapping.
	if hasOp {
		return true
	}
	return strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "+")
}

func canStartArithRegion(s string, i int) bool {
	if i >= len(s) {
		return false
	}
	c := s[i]
	if c >= '0' && c <= '9' || c == '(' || c == '.' {
		return true
	}
	if c == '+' || c == '-' {
		if i > 0 {
			prev := s[i-1]
			// Don't treat hyphen inside account names / words as unary minus.
			if isIdentByte(prev) || prev == ':' {
				return false
			}
		}
		if i+1 < len(s) {
			n := s[i+1]
			return n >= '0' && n <= '9' || n == '.' || n == '('
		}
	}
	return false
}

// rewriteArithmeticRegionsStrict finds maximal arithmetic regions (digits/ops/parens)
// that contain binary operators or parentheses and evaluates each with standard
// precedence. Incomplete or illegal expressions return an error (no left-to-right
// regex rewrite, no silent fallback). Cost braces `{...}` / `{{...}}` are copied
// verbatim so cost dates are never treated as subtraction.
func rewriteArithmeticRegionsStrict(value string) (string, error) {
	if value == "" {
		return value, nil
	}
	var b strings.Builder
	b.Grow(len(value))
	i := 0
	for i < len(value) {
		if value[i] == '{' {
			end := skipBalancedBraces(value, i)
			b.WriteString(value[i:end])
			i = end
			continue
		}
		if !canStartArithRegion(value, i) {
			b.WriteByte(value[i])
			i++
			continue
		}
		start := i
		i = extendArithRegion(value, start)
		end := i
		for end > start && (value[end-1] == ' ' || value[end-1] == '\t') {
			end--
		}
		region := value[start:end]
		if containsBinaryArithmetic(region) || strings.ContainsAny(region, "()") {
			out, err := evalArithmeticExpression(strings.TrimSpace(region))
			if err != nil {
				return "", fmt.Errorf("arithmetic %q: %w", strings.TrimSpace(region), err)
			}
			b.WriteString(formatAmountLikeDecimal(out, strings.TrimSpace(region)))
		} else {
			b.WriteString(region)
		}
		for end < i {
			b.WriteByte(value[end])
			end++
		}
	}
	return b.String(), nil
}

func skipBalancedBraces(s string, start int) int {
	if start >= len(s) || s[start] != '{' {
		return start
	}
	depth := 0
	for i := start; i < len(s); i++ {
		switch s[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i + 1
			}
		}
	}
	return len(s)
}

// extendArithRegion advances past a numeric/operator/paren run starting at start.
// 'e'/'E' are only accepted as scientific-notation exponents after digits.
func extendArithRegion(s string, start int) int {
	i := start
	sawDigit := false
	for i < len(s) {
		c := s[i]
		switch {
		case c >= '0' && c <= '9':
			sawDigit = true
			i++
		case c == '.' || c == '+' || c == '-' || c == '*' || c == '/' || c == '(' || c == ')':
			i++
		case c == ' ' || c == '\t':
			i++
		case (c == 'e' || c == 'E') && sawDigit:
			i++
			if i < len(s) && (s[i] == '+' || s[i] == '-') {
				i++
			}
			exp := false
			for i < len(s) && s[i] >= '0' && s[i] <= '9' {
				exp = true
				i++
			}
			if !exp {
				// Lone 'e' is not an exponent — stop before it.
				i--
				if s[i] == 'e' || s[i] == 'E' {
					return i
				}
			}
			sawDigit = true
		default:
			return i
		}
	}
	return i
}

// validatePostingAmountTokens ensures amount fields on posting-like lines parse as money.
// Supports Beancount-ish tails: [amount] [commodity] [{cost}] [@/@ @ price] [; comment],
// elided amounts, and commodities that contain digits. It is deliberately token-aware
// rather than joining the entire tail as one numeric string.
func validatePostingAmountTokens(line string) error {
	s := strings.TrimSpace(line)
	if s == "" {
		return nil
	}
	if i := strings.Index(s, ";"); i >= 0 {
		s = strings.TrimSpace(s[:i])
	}
	if s == "" {
		return nil
	}
	account, rest := splitPostingAccount(s)
	_ = account
	rest = strings.TrimSpace(rest)
	if rest == "" {
		return nil // elided amount
	}
	amount, rest, err := takePostingAmountToken(rest)
	if err != nil {
		return fmt.Errorf("invalid posting amount in %q: %w", line, err)
	}
	if amount == "" && rest != "" && !strings.HasPrefix(rest, "{") && !strings.HasPrefix(rest, "@") {
		// Non-numeric first token (e.g. "nonsense CNY") is an attempted amount,
		// not a bare commodity. Beancount units require a number when present.
		tok, after := takePostingCommodity(rest)
		if tok == "" {
			tok, after = firstPostingField(rest)
		}
		if tok != "" {
			amount = tok
			rest = after
		}
	}
	if amount != "" {
		if _, err := ParseAmountDecimal(amount, ""); err != nil {
			return fmt.Errorf("invalid posting amount %q in %q: %w", amount, line, err)
		}
	}
	rest = strings.TrimSpace(rest)
	_, rest = takePostingCommodity(rest)
	rest = strings.TrimSpace(rest)
	rest, err = skipPostingCostSpec(rest)
	if err != nil {
		return fmt.Errorf("invalid posting cost in %q: %w", line, err)
	}
	rest = strings.TrimSpace(rest)
	rest, err = skipPostingPriceSpec(rest)
	if err != nil {
		return fmt.Errorf("invalid posting price in %q: %w", line, err)
	}
	rest = strings.TrimSpace(rest)
	if rest != "" {
		// Unexpected trailing tokens — reject money-like leftovers, else ignore.
		if mightBeMoneyText(rest) {
			return fmt.Errorf("invalid posting trailing %q in %q", rest, line)
		}
	}
	return nil
}

func splitPostingAccount(line string) (account, rest string) {
	trimmed := strings.TrimLeftFunc(line, unicode.IsSpace)
	lead := line[:len(line)-len(trimmed)]
	end := strings.IndexFunc(trimmed, unicode.IsSpace)
	if end < 0 {
		return lead + trimmed, ""
	}
	tok := trimmed[:end]
	if strings.Contains(tok, ":") {
		return lead + tok, trimmed[end:]
	}
	return "", line
}

func takePostingAmountToken(rest string) (amount, remaining string, err error) {
	rest = strings.TrimSpace(rest)
	if rest == "" {
		return "", "", nil
	}
	// Cost/price may appear with elided incomplete amount.
	if strings.HasPrefix(rest, "{") || strings.HasPrefix(rest, "@") {
		return "", rest, nil
	}
	i := 0
	if rest[0] == '+' || rest[0] == '-' {
		i = 1
	}
	if i >= len(rest) {
		return "", rest, fmt.Errorf("expected amount")
	}
	start := 0
	// Number / arithmetic already reduced to a plain decimal token.
	sawDigit := false
	for i < len(rest) {
		c := rest[i]
		switch {
		case c >= '0' && c <= '9':
			sawDigit = true
			i++
		case c == '.':
			i++
		case (c == 'e' || c == 'E') && sawDigit:
			i++
			if i < len(rest) && (rest[i] == '+' || rest[i] == '-') {
				i++
			}
			for i < len(rest) && rest[i] >= '0' && rest[i] <= '9' {
				i++
			}
		default:
			goto done
		}
	}
done:
	if !sawDigit {
		lower := strings.ToLower(rest)
		for _, bad := range []string{"nan", "inf", "+inf", "-inf", "infinity", "+infinity", "-infinity"} {
			if strings.HasPrefix(lower, bad) {
				tok := rest[:len(bad)]
				if len(rest) == len(bad) || unicode.IsSpace(rune(rest[len(bad)])) {
					return tok, strings.TrimSpace(rest[len(bad):]), nil
				}
			}
		}
		return "", rest, nil
	}
	return rest[start:i], rest[i:], nil
}

func takePostingCommodity(rest string) (commodity, remaining string) {
	rest = strings.TrimSpace(rest)
	if rest == "" || rest[0] == '{' || rest[0] == '@' || rest[0] == ';' {
		return "", rest
	}
	i := 0
	for i < len(rest) {
		c := rest[i]
		if c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_' || c == '.' || c == '-' {
			i++
			continue
		}
		break
	}
	if i == 0 {
		return "", rest
	}
	// Commodity must start with a letter.
	if c := rest[0]; (c < 'A' || c > 'Z') && (c < 'a' || c > 'z') {
		return "", rest
	}
	return rest[:i], rest[i:]
}

func firstPostingField(rest string) (tok, remaining string) {
	rest = strings.TrimSpace(rest)
	if rest == "" {
		return "", ""
	}
	end := strings.IndexFunc(rest, unicode.IsSpace)
	if end < 0 {
		return rest, ""
	}
	return rest[:end], rest[end:]
}

func skipPostingCostSpec(rest string) (string, error) {
	rest = strings.TrimSpace(rest)
	if !strings.HasPrefix(rest, "{") {
		return rest, nil
	}
	end := skipBalancedBraces(rest, 0)
	if end <= 0 || (end > 0 && rest[end-1] != '}') {
		return rest, fmt.Errorf("unclosed cost braces")
	}
	return rest[end:], nil
}

func skipPostingPriceSpec(rest string) (string, error) {
	rest = strings.TrimSpace(rest)
	if !strings.HasPrefix(rest, "@") {
		return rest, nil
	}
	if strings.HasPrefix(rest, "@@") {
		rest = strings.TrimSpace(rest[2:])
	} else {
		rest = strings.TrimSpace(rest[1:])
	}
	amount, rest, err := takePostingAmountToken(rest)
	if err != nil {
		return rest, err
	}
	if amount == "" {
		return rest, fmt.Errorf("expected price amount")
	}
	if _, err := ParseAmountDecimal(amount, ""); err != nil {
		return rest, err
	}
	_, rest = takePostingCommodity(rest)
	return rest, nil
}

func mightBeMoneyText(s string) bool {
	lower := strings.ToLower(strings.TrimSpace(s))
	switch lower {
	case "nan", "inf", "+inf", "-inf", "infinity", "+infinity", "-infinity":
		return true
	}
	if containsBinaryArithmetic(s) {
		return true
	}
	for _, r := range s {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}
