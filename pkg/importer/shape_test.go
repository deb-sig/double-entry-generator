package importer

import (
	"strings"
	"testing"

	"github.com/deb-sig/double-entry-generator/v2/pkg/reader"
)

func shapeRows(t *testing.T, profileYAML string, rows [][]string) ([]string, [][]string) {
	t.Helper()
	profile := loadProfileYAML(t, profileYAML)
	headers, out, err := shapeTable(profile, reader.Table{Rows: rows})
	if err != nil {
		t.Fatal(err)
	}
	return headers, out
}

// Alipay dropped one disclaimer line in 2026 (#236); counting leading rows
// then swallows the first transaction. Anchors do not care.
func TestShapeLocateHeaderByAnchorIgnoresPreambleLength(t *testing.T) {
	profile := `
shape:
  - locateHeader: { anchor: [交易时间, 金额], scanRows: 20 }
`
	short := [][]string{{"支付宝账单"}, {"交易时间", "交易对方", "金额"}, {"2026-01-01", "A", "1"}}
	long := [][]string{{"支付宝账单"}, {"仅展示当前交易"}, nil, {"交易时间", "交易对方", "金额"}, {"2026-01-01", "A", "1"}}
	for _, rows := range [][][]string{short, long} {
		headers, out := shapeRows(t, profile, rows)
		if strings.Join(headers, ",") != "交易时间,交易对方,金额" {
			t.Fatalf("headers = %q", headers)
		}
		if len(out) != 1 || out[0][0] != "2026-01-01" {
			t.Fatalf("rows = %q", out)
		}
	}
}

func TestShapeLocateHeaderReportsMissingAnchor(t *testing.T) {
	profile := loadProfileYAML(t, `
shape:
  - locateHeader: { anchor: 不存在的列 }
`)
	_, _, err := shapeTable(profile, reader.Table{Rows: [][]string{{"a", "b"}}})
	if err == nil || !strings.Contains(err.Error(), "不存在的列") {
		t.Fatalf("err = %v", err)
	}
}

func TestShapeDropMatchingAndDropIf(t *testing.T) {
	headers, out := shapeRows(t, `
shape:
  - dropMatching: '^(共计|合计|---)'
  - locateHeader: { anchor: 状态 }
  - dropIf: '<状态> ~ "失败" || <金额>.number == 0'
`, [][]string{
		{"导出说明"},
		{"日期", "金额", "状态"},
		{"2026-01-01", "10", "成功"},
		{"2026-01-02", "20", "还款失败"},
		{"2026-01-03", "0", "成功"},
		{"合计", "30", ""},
	})
	if strings.Join(headers, ",") != "日期,金额,状态" {
		t.Fatalf("headers = %q", headers)
	}
	if len(out) != 1 || out[0][0] != "2026-01-01" {
		t.Fatalf("rows = %q", out)
	}
}

// hxsec exports one fill as an amount-only line plus a quantity/price line
// sharing the contract and fill numbers. Today's rules ignore one line and
// recompute on the other; merge makes them one record.
func TestShapeMergeSplitFills(t *testing.T) {
	_, out := shapeRows(t, `
template:
  sourceHeaders: [合同号, 成交号, 成交金额, 成交数量, 成交价格, 手续费, 备注]
shape:
  - merge:
      key: [<合同号>, <成交号>]
      take: { 成交金额: first-nonzero, 成交数量: first-nonzero, 成交价格: first-nonzero, 手续费: sum, 备注: join }
`, [][]string{
		{"合同号", "成交号", "成交金额", "成交数量", "成交价格", "手续费", "备注"},
		{"C1", "F1", "1000.00", "0", "0", "5.00", "买入"},
		{"C2", "F2", "0", "50", "20.10", "0.10", "卖出"},
		{"C1", "F1", "0", "100", "10.00", "0.30", "买入"},
		{"", "", "7.00", "0", "0", "0", "无合同号的行保持独立"},
	})
	want := [][]string{
		{"C1", "F1", "1000.00", "100", "10.00", "5.30", "买入"},
		{"C2", "F2", "0", "50", "20.10", "0.10", "卖出"},
		{"", "", "7.00", "0", "0", "0", "无合同号的行保持独立"},
	}
	if len(out) != len(want) {
		t.Fatalf("rows = %q", out)
	}
	for i := range want {
		if strings.Join(out[i], "|") != strings.Join(want[i], "|") {
			t.Fatalf("row %d = %q, want %q", i, out[i], want[i])
		}
	}
}

func TestShapeMergeRejectsUnknownTakeMode(t *testing.T) {
	profile := loadProfileYAML(t, `
template:
  sourceHeaders: [a]
shape:
  - merge: { key: <a>, take: { a: median } }
`)
	_, _, err := shapeTable(profile, reader.Table{Rows: [][]string{{"a"}, {"1"}}})
	if err == nil || !strings.Contains(err.Error(), "median") {
		t.Fatalf("err = %v", err)
	}
}

// One WeChat withdrawal line carries both the transfer and its fee. Split
// turns it into two records so each can be a plain two-leg transaction.
func TestShapeSplitFeeIntoOwnRecord(t *testing.T) {
	_, out := shapeRows(t, `
template:
  sourceHeaders: [交易单号, 交易类型, 金额, 备注]
shape:
  - split:
      when: '<交易类型> == "零钱提现" && <备注> ~ "服务费"'
      into:
        - {}
        - { 交易类型: 手续费, 金额: '<备注>.extract("服务费.?([.0-9]+)")' }
`, [][]string{
		{"交易单号", "交易类型", "金额", "备注"},
		{"T1", "零钱提现", "100.00", "服务费¥0.10"},
		{"T2", "商户消费", "5.00", ""},
	})
	want := [][]string{
		{"T1", "零钱提现", "100.00", "服务费¥0.10"},
		{"T1", "手续费", "0.10", "服务费¥0.10"},
		{"T2", "商户消费", "5.00", ""},
	}
	if len(out) != len(want) {
		t.Fatalf("rows = %q", out)
	}
	for i := range want {
		if strings.Join(out[i], "|") != strings.Join(want[i], "|") {
			t.Fatalf("row %d = %q, want %q", i, out[i], want[i])
		}
	}
}

// Without a locateHeader step the legacy template fields still decide
// where the header is, so shape can be adopted one step at a time.
func TestShapeFallsBackToLegacyHeaderFields(t *testing.T) {
	headers, out := shapeRows(t, `
template:
  skipLeadingRows: 1
shape:
  - dropIf: '<b> == "x"'
`, [][]string{{"说明"}, {"a", "b"}, {"1", "x"}, {"2", "y"}})
	if strings.Join(headers, ",") != "a,b" || len(out) != 1 || out[0][0] != "2" {
		t.Fatalf("headers = %q rows = %q", headers, out)
	}
}

func TestShapeEndToEndImport(t *testing.T) {
	profile := loadProfileYAML(t, `
schema: https://deg.dev/template-profile/v2
id: shaped
reader: { format: csv }
shape:
  - locateHeader: { anchor: 日期 }
  - dropIf: '<状态> == "失败"'
template:
  defaultCurrency: CNY
templateRules:
  - id: base
    actions:
      date: <日期>
      payee: <对方>
      amount: <金额>.number
      from: Assets:Cash
      to: Expenses:FIXME
`)
	data := []byte("账单\n随便一行\n日期,对方,金额,状态\n2026-01-01,A,-1.00,成功\n2026-01-02,B,-2.00,失败\n")
	rows, err := ParseBytes(profile, "bill.csv", data)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Raw["对方"] != "A" {
		t.Fatalf("rows = %+v", rows)
	}
}
