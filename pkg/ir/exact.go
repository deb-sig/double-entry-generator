package ir

// Decimal 收尾（2026-09-27）：把「精确金额」的构造收敛到一处。
//
// 背景：引擎侧早已按 ExactMoney 求值（`order.ExactMoney` 权威、
// `order.Money`(float64) 仅作遗留视图），但**21 个 provider 没有一个写 ExactMoney** ——
// 全都在 `Money: float64(x)/100.0` 这条路上丢精度（USDT/币量 8 位小数尤其明显）。
//
// 迁移口径：能拿到**整数分/整数单位或原文**的 provider 一律改走下面的构造函数，
// 不再从 float64 反推；拿不到的（原文在解析时已被 float64 吃掉）需要先改那个
// provider 的解析保留原文，属于逐 provider 的活，见 exact_money_gate_test.go 的清单。

// ExactFromUnits 用「整数单位 + scale」构造精确金额，不经过 float64。
// 例：ExactFromUnits(12345, 2) == 123.45。
func ExactFromUnits(units int64, scale uint32) *Decimal {
	var d Decimal
	d.units.SetInt64(units)
	d.scale = scale
	return &d
}

// ExactFromCents 「分」→ 元（scale 2）。
func ExactFromCents(cents int64) *Decimal { return ExactFromUnits(cents, 2) }

// ExactFromText 从**原文**构造精确金额（已经是字符串的 provider 用这个）。
// 文本清洗（货币符号、千分位）交给 importer.CleanAmount。
func ExactFromText(text string) (*Decimal, error) {
	d, err := ParseDecimal(text)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// SetExactFromCents 同时写权威值和 legacy 视图，供不方便用结构体字面量的地方调用。
func (o *Order) SetExactFromCents(cents int64) {
	d := ExactFromCents(cents)
	o.ExactMoney = d
	o.Money = d.Float64Approx()
}
