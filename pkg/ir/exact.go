package ir

import "strings"

// Decimal 收尾（2026-09-27）：把「精确金额」的构造收敛到一处。
//
// 背景：引擎侧按 ExactMoney 求值（`order.ExactMoney` 权威、`order.Money`(float64) 为遗留视图）。
//
// **2026-09-27 实测纠正**：provider 的 `Money`(float64) 在现有两条实际流程里
// **都不是权威来源** —— 引擎会从**原文**（列映射的金额列 / 规则里的 `amount:` 动作）
// 重新构造 ExactMoney。用「迁移前 / 迁移后」两个二进制跑
//   ① profile 流程 19 家模板 + ② provider 流程（oklink/huobi 例子）
// 输出**逐字节一致**；全仓 `order.Money` 只剩两处用途：赋值本身 + `inferType` 方向推断。
// 所以本文件的构造函数是**防退化**用（挡住未来新增的丢精度实现），不是修可见 bug。
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

// ExactFromTextOrNil 是 ExactFromText 的「拿不到就 nil」版本，给 provider 用：
// 原文解析失败时退回让引擎走旧的 float64 视图，而不是让整次导入失败。
func ExactFromTextOrNil(text string) *Decimal {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	d, err := ParseDecimal(text)
	if err != nil {
		return nil
	}
	return &d
}
