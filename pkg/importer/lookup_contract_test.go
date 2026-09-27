package importer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
)

// Lookup contract: conditions use five lowercase logical fallbacks with literal
// custom columns; action <col> refs are Raw-only (Mirato __from_column family).
//
// 顺序语义（2026-09-27 定案）：raw 同名列 > 当前 order 槽；要账单原值用 original_*/imported_*。

func TestLookupConditionFiveLogicalsAndLiteralCustom(t *testing.T) {
	row := Row{
		Payee:     "LogicalPayee",
		Narration: "LogicalNarr",
		Amount:    "5.00",
		Currency:  "CNY",
		Date:      "2026-09-01",
		Raw: map[string]string{
			"payee":     "RawPayee",
			"商家":        "中文商家",
			"empty":     "",
			"raw.payee": "LiteralRawDot",
		},
	}
	order := ir.Order{Peer: "Changed", Item: "Mutated"}

	cases := []struct {
		when string
		want bool
	}{
		// Raw same-name wins over logical.
		{`payee == "RawPayee"`, true},
		{`payee == "LogicalPayee"`, false},
		// 缺失的标准字段回落逻辑槽 —— 但读的是**当前 order 值**（顺序语义，2026-09-27 定案）：
		// order.Narration 已被改成 "Mutated"，所以逻辑槽看到的是 Mutated 而不是 row 原值。
		{`narration == "LogicalNarr"`, false},
		{`narration == "Mutated"`, true},
		// 要读账单原值必须用不可变字段（对标 Actual Budget 的 imported payee）
		{`original_narration == "LogicalNarr"`, true},
		{`original_narration == "Mutated"`, false},
		{`original_payee == "LogicalPayee"`, true}, // 不被 raw 同名列遮蔽，也不被改写影响
		{`amount == "5.00"`, true},
		{`currency == "CNY"`, true},
		{`date == "2026-09-01"`, true},
		// Custom / namespaced columns are literal identity.
		{`<商家> == "中文商家"`, true},
		{`<raw.payee> == "LiteralRawDot"`, true},
		{`<raw.payee> == "RawPayee"`, false}, // must NOT strip raw. namespace
		{`raw.payee == "LiteralRawDot"`, true},
		// Present empty stays empty (no logical fallback).
		{`<empty> == ""`, true},
		// Missing custom → "".
		{`<missing> == ""`, true},
		{`missing == ""`, true},
		// No case fold / peer / item aliases.
		{`<Payee> == "LogicalPayee"`, false},
		{`<Payee> == "RawPayee"`, false},
		{`peer == "LogicalPayee"`, false},
		{`item == "LogicalNarr"`, false},
		{`<metadata.payee> == "RawPayee"`, false},
	}
	for _, tc := range cases {
		ok, err := evalWhen(tc.when, row, order)
		if err != nil {
			t.Fatalf("when %q: %v", tc.when, err)
		}
		if ok != tc.want {
			t.Fatalf("when %q = %v, want %v", tc.when, ok, tc.want)
		}
	}
}

func TestLookupActionAngleRefsAreRawOnly(t *testing.T) {
	row := Row{
		Payee:    "LogicalPayee",
		Amount:   "5.00",
		Currency: "CNY",
		Raw: map[string]string{
			"payee":    "RawPayee",
			"empty":    "",
			"商家":       "Original",
			"raw.payee": "LiteralRawDot",
		},
	}
	order := ir.Order{Peer: "Changed", Currency: "EUR"}

	cases := []struct {
		expr string
		want string
	}{
		{`<payee>`, "RawPayee"},
		{`<amount>`, ""}, // no raw amount key — must NOT use logical 5.00
		{`<currency>`, ""},
		{`<missing>`, ""},
		{`<empty>`, ""},
		{`<商家>`, "Original"},
		{`<商家>.replace("Original","Replaced")`, "Replaced"},
		{`<raw.payee>`, "LiteralRawDot"}, // literal key, not namespace
		{`<Payee>`, ""},                 // no case fold into payee
	}
	for _, tc := range cases {
		got := resolveActionValue(tc.expr, row, order)
		if got != tc.want {
			t.Fatalf("action %s = %q, want %q", tc.expr, got, tc.want)
		}
	}

	// DEG-native bracket refs still see logical amount for transfer legs.
	if got := resolveActionValue(`[amount]`, row, order); got != "5.00" {
		t.Fatalf("bracket [amount] logical fallback broken: %q", got)
	}
}

func TestLookupDynamicActionsOnChineseCSVMatchMiratoRawOnly(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "bill.csv")
	// Chinese headers only — mapped to logical payee/amount/currency; no English raw keys.
	body := "交易时间,交易对方,商品,收/支,金额,币种,支付方式\n" +
		"2026-09-01 10:00:00,Original,午餐,支出,5.00,CNY,余额\n"
	if err := os.WriteFile(csvPath, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	profile := testProfile()
	profile.Schema = "https://deg.dev/template-profile/v2"
	profile.Template.Columns.Date = "交易时间"
	profile.Template.Columns.Payee = "交易对方"
	profile.Template.Columns.Narration = "商品"
	profile.Template.Columns.Amount = "金额"
	profile.Template.Columns.Currency = "币种"
	profile.Template.Columns.Type = "收/支"
	profile.Template.DefaultMinus = "Assets:Cash"
	profile.Template.DefaultPlus = "Expenses:Food"
	profile.RequiredCapabilities = []string{"when.literalFieldLookup", "actions.rawColumnRef"}
	enabled := true
	profile.PersonalRules = []Rule{
		{
			Enabled: &enabled,
			Actions: Actions{
				Payee:     `"Changed"`,
				Amount:    `99.00`,
				Currency:  `EUR`,
				From:      TransferSide{Account: "Assets:Cash"},
				To:        TransferSide{Account: "Expenses:Food"},
				Narration: `<payee>`,
				Metadata: map[string]string{
					"origamount":   `<amount>`,
					"origcurrency": `<currency>`,
					"replace":      `<payee>.replace("Original","Replaced")`,
					"missing":      `<missing>`,
				},
			},
		},
		{
			Enabled: &enabled,
			When:    `<raw.payee> =~ "^$"`,
			Actions: Actions{Tag: `"rawmissing"`},
		},
	}
	out, err := ImportFile(profile, csvPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Orders) != 1 {
		t.Fatalf("%#v", out.Orders)
	}
	o := out.Orders[0]
	if o.Peer != "Changed" {
		t.Fatalf("payee=%q", o.Peer)
	}
	if o.Item != "" {
		t.Fatalf("narration must be raw-only empty, got %q", o.Item)
	}
	if o.Metadata["origamount"] != "" || o.Metadata["origcurrency"] != "" || o.Metadata["replace"] != "" {
		t.Fatalf("dynamic metadata must be empty when English raw cols absent: %#v", o.Metadata)
	}
	if o.Metadata["missing"] != "" {
		t.Fatalf("missing col must be empty string, got %#v", o.Metadata["missing"])
	}
	joined := strings.Join(o.Tags, ",")
	if !strings.Contains(joined, "rawmissing") {
		t.Fatalf("literal raw.payee missing must match empty regex, tags=%v", o.Tags)
	}
}

func TestLookupRawPresentEnglishColumnsStayRaw(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "bill.csv")
	body := "交易时间,交易对方,商品,收/支,金额,币种,支付方式,payee,amount,currency\n" +
		"2026-09-01 10:00:00,Original,午餐,支出,5.00,CNY,余额,RawOriginal,7.00,USD\n"
	if err := os.WriteFile(csvPath, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	profile := testProfile()
	profile.Schema = "https://deg.dev/template-profile/v2"
	profile.Template.Columns.Date = "交易时间"
	profile.Template.Columns.Payee = "交易对方"
	profile.Template.Columns.Narration = "商品"
	profile.Template.Columns.Amount = "金额"
	profile.Template.Columns.Currency = "币种"
	profile.Template.Columns.Type = "收/支"
	profile.Template.DefaultMinus = "Assets:Cash"
	profile.Template.DefaultPlus = "Expenses:Food"
	enabled := true
	profile.PersonalRules = []Rule{
		{
			Enabled: &enabled,
			Actions: Actions{
				Payee:     `"Changed"`,
				Amount:    `99.00`,
				Currency:  `EUR`,
				From:      TransferSide{Account: "Assets:Cash"},
				To:        TransferSide{Account: "Expenses:Food"},
				Narration: `<payee>`,
				Metadata: map[string]string{
					"origamount":   `<amount>`,
					"origcurrency": `<currency>`,
					"replace":      `<payee>.replace("Original","Replaced")`,
				},
			},
		},
		{
			Enabled: &enabled,
			When:    `<raw.payee> =~ "^$"`,
			Actions: Actions{Tag: `"rawmissing"`},
		},
	}
	out, err := ImportFile(profile, csvPath)
	if err != nil {
		t.Fatal(err)
	}
	o := out.Orders[0]
	if o.Item != "RawOriginal" {
		t.Fatalf("narration=%q", o.Item)
	}
	if o.Metadata["origamount"] != "7.00" || o.Metadata["origcurrency"] != "USD" {
		t.Fatalf("metadata=%#v", o.Metadata)
	}
	if o.Metadata["replace"] != "RawReplaced" {
		// replace operates on RawOriginal: "Original" substring → "Replaced" => RawReplaced
		t.Fatalf("replace=%q", o.Metadata["replace"])
	}
	joined := strings.Join(o.Tags, ",")
	// Literal column name "raw.payee" is still absent — Mirato hits rawmissing.
	if !strings.Contains(joined, "rawmissing") {
		t.Fatalf("literal raw.payee key absent must match empty regex like Mirato; tags=%v", o.Tags)
	}
}
