package importer

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Sequence contract (Mirato baseline).
//
// Observation layer note (2026-09-23 lookup repair):
// V2 ImportFile writes legs to Order.Postings via renderV2Postings; V1
// PlusAccount is not mirrored. Independent review confirmed three earlier
// failures checked the wrong representation layer. This file asserts the
// full posting list (account + amount + currency), which is stricter than
// a mere account-contains check. Original red log preserved at
// workspace/deg-p1-lookup-repair/original-sequence-failure.log.

var postingLineRE = regexp.MustCompile(`^(\S+)\s+(-?[0-9]+(?:\.[0-9]+)?)\s+(\S+)\s*$`)

type expectPosting struct {
	Account  string
	Amount   string
	Currency string
}

func parsePostingLine(line string) (expectPosting, bool) {
	m := postingLineRE.FindStringSubmatch(strings.TrimSpace(line))
	if m == nil {
		return expectPosting{}, false
	}
	return expectPosting{Account: m[1], Amount: m[2], Currency: m[3]}, true
}

func assertPostingsEqual(t *testing.T, gotLines []string, want []expectPosting) {
	t.Helper()
	got := make([]expectPosting, 0, len(gotLines))
	for _, line := range gotLines {
		p, ok := parsePostingLine(line)
		if !ok {
			t.Fatalf("unparseable posting line %q (full=%v)", line, gotLines)
		}
		got = append(got, p)
	}
	if len(got) != len(want) {
		t.Fatalf("posting count: got %#v want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("posting[%d]: got %#v want %#v (full got=%#v)", i, got[i], want[i], got)
		}
	}
}

func TestSequenceRawFieldWinsOverLogicalPayee(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "bill.csv")
	body := "交易时间,交易对方,商品,收/支,金额,支付方式,payee\n" +
		"2026-05-21 10:30:00,映射B,午餐,支出,7.00,余额,原始A\n"
	if err := os.WriteFile(csvPath, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	profile := testProfile()
	profile.Schema = "https://deg.dev/template-profile/v2"
	profile.Template.DefaultMinus = "Assets:Cash"
	profile.Template.DefaultPlus = "Expenses:Misc"
	enabled := true
	profile.PersonalRules = []Rule{
		{
			ID:      "raw-win",
			Enabled: &enabled,
			When:    `payee == "原始A"`,
			Actions: Actions{
				To:   TransferSide{Account: "Expenses:Food"},
				From: TransferSide{Account: "Assets:Cash"},
				Tag:  "rawwin",
			},
		},
		{
			ID:      "mapped-trap",
			Enabled: &enabled,
			When:    `payee == "映射B"`,
			Actions: Actions{
				To:  TransferSide{Account: "Expenses:ShouldNot"},
				Tag: "mappedtrap",
			},
		},
	}
	out, err := ImportFile(profile, csvPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Orders) != 1 {
		t.Fatalf("want 1 order, got %#v", out.Orders)
	}
	o := out.Orders[0]
	assertPostingsEqual(t, postingLines(o), []expectPosting{
		{Account: "Expenses:Food", Amount: "7.00", Currency: "CNY"},
		{Account: "Assets:Cash", Amount: "-7.00", Currency: "CNY"},
	})
	joined := strings.Join(o.Tags, ",")
	if !strings.Contains(joined, "rawwin") || strings.Contains(joined, "mappedtrap") {
		t.Fatalf("tags=%v", o.Tags)
	}
}

func TestSequenceMissingAndEmptyFieldsAreEmptyString(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "bill.csv")
	// 空列 present empty; 缺列 absent from header.
	body := "交易时间,交易对方,商品,收/支,金额,支付方式,空列\n" +
		"2026-05-21 10:30:00,空列店,午餐,支出,8.00,余额,\n"
	if err := os.WriteFile(csvPath, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	profile := testProfile()
	profile.Schema = "https://deg.dev/template-profile/v2"
	enabled := true
	profile.PersonalRules = []Rule{
		{
			Enabled: &enabled,
			When:    `<空列> == "" && payee == "空列店"`,
			Actions: Actions{Tag: "emptycol", To: TransferSide{Account: "Expenses:Food"}},
		},
		{
			Enabled: &enabled,
			When:    `<缺列> == "" && payee == "空列店"`,
			Actions: Actions{Tag: "missingcol"},
		},
	}
	out, err := ImportFile(profile, csvPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Orders) != 1 {
		t.Fatalf("%#v", out.Orders)
	}
	joined := strings.Join(out.Orders[0].Tags, ",")
	if !strings.Contains(joined, "emptycol") || !strings.Contains(joined, "missingcol") {
		t.Fatalf("missing/empty must be \"\", tags=%v", out.Orders[0].Tags)
	}
	// Bare missing identifier must not become the field-name literal.
	ok, err := evalWhen(`缺列 == ""`, Row{Payee: "x", Raw: map[string]string{}}, out.Orders[0])
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal(`bare missing field must compare as "" not "缺列"`)
	}
}

// 顺序语义（2026-09-27 定案，对齐 Actual Budget / Firefly III）：
// 规则按序执行，后面的条件看到前面**改写后**的值；要原值必须用 original_* 显式引用。
// 本用例的 trap 规则因此**会**命中（旧断言假设它不命中，是错的）。
func TestSequenceConditionsSeeMutatedOrder(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "bill.csv")
	body := "交易时间,交易对方,商品,收/支,金额,支付方式\n" +
		"2026-05-21 10:30:00,原始条件店,午餐,支出,5.00,余额\n"
	if err := os.WriteFile(csvPath, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	profile := testProfile()
	profile.Schema = "https://deg.dev/template-profile/v2"
	enabled := true
	profile.PersonalRules = []Rule{
		{
			Enabled: &enabled,
			When:    `payee == "原始条件店"`,
			Actions: Actions{Payee: `"已改名店"`, To: TransferSide{Account: "Expenses:Temp"}},
		},
		{
			Enabled: &enabled,
			When:    `payee == "已改名店"`,
			Actions: Actions{To: TransferSide{Account: "Expenses:ShouldNot"}, Tag: "trap"},
		},
		{
			Enabled: &enabled,
			When:    `original_payee == "原始条件店"`,
			Actions: Actions{To: TransferSide{Account: "Expenses:Food"}, Tag: "orig"},
		},
	}
	out, err := ImportFile(profile, csvPath)
	if err != nil {
		t.Fatal(err)
	}
	o := out.Orders[0]
	if o.Peer != "已改名店" {
		t.Fatalf("payee rewrite: %q", o.Peer)
	}
	// trap 命中（条件看到改写值）→ ShouldNot；orig 用 original_payee 也命中 → 后写的 Food 胜
	assertPostingsEqual(t, postingLines(o), []expectPosting{
		{Account: "Expenses:Food", Amount: "5.00", Currency: "CNY"},
	})
	joined := strings.Join(o.Tags, ",")
	if !strings.Contains(joined, "trap") || !strings.Contains(joined, "orig") {
		t.Fatalf("两条规则都该命中（trap=看到新值、orig=读原值），tags=%v", o.Tags)
	}
}

func TestSequenceStickyIgnoreAndEnabledAndTemplateScope(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "bill.csv")
	body := "交易时间,交易对方,商品,收/支,金额,支付方式\n" +
		"2026-05-21 10:30:00,粘性忽略店,午餐,支出,3.00,余额\n" +
		"2026-05-21 10:31:00,禁用顺序店,午餐,支出,4.00,余额\n" +
		"2026-05-21 10:32:00,作用域顺序店,午餐,支出,6.00,余额\n"
	if err := os.WriteFile(csvPath, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	profile := testProfile()
	profile.ID = "native-sequence"
	profile.Schema = "https://deg.dev/template-profile/v2"
	enabled := true
	disabled := false
	profile.PersonalRules = []Rule{
		{Enabled: &enabled, When: `payee == "粘性忽略店"`, Actions: Actions{Ignore: true}},
		{Enabled: &enabled, When: `payee == "粘性忽略店"`, Actions: Actions{
			Ignore: false, To: TransferSide{Account: "Expenses:ShouldNot"},
		}},
		{Enabled: &disabled, When: `payee == "禁用顺序店"`, Actions: Actions{To: TransferSide{Account: "Expenses:ShouldNot"}}},
		{Enabled: &enabled, When: `payee == "禁用顺序店"`, Actions: Actions{To: TransferSide{Account: "Expenses:Food"}}},
		{Enabled: &enabled, TemplateID: "native-sequence", When: `payee == "作用域顺序店"`,
			Actions: Actions{To: TransferSide{Account: "Expenses:Food"}, Tag: "inscope"}},
		{Enabled: &enabled, TemplateID: "other-template", When: `payee == "作用域顺序店"`,
			Actions: Actions{To: TransferSide{Account: "Expenses:ShouldNot"}, Tag: "leak"}},
	}
	out, err := ImportFile(profile, csvPath)
	if err != nil {
		t.Fatal(err)
	}
	var peers []string
	for _, o := range out.Orders {
		peers = append(peers, o.Peer)
		if o.Peer == "禁用顺序店" {
			assertPostingsEqual(t, postingLines(o), []expectPosting{
				{Account: "Expenses:Food", Amount: "4.00", Currency: "CNY"},
			})
		}
		if o.Peer == "作用域顺序店" {
			assertPostingsEqual(t, postingLines(o), []expectPosting{
				{Account: "Expenses:Food", Amount: "6.00", Currency: "CNY"},
			})
			joined := strings.Join(o.Tags, ",")
			if strings.Contains(joined, "leak") || !strings.Contains(joined, "inscope") {
				t.Fatalf("scope tags=%v", o.Tags)
			}
		}
	}
	for _, p := range peers {
		if p == "粘性忽略店" {
			t.Fatalf("sticky ignore must drop order, peers=%v", peers)
		}
	}
}

func TestSequenceAppendReplaceFromToOrder(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "bill.csv")
	body := "交易时间,交易对方,商品,收/支,金额,支付方式\n" +
		"2026-05-21 10:30:00,腿顺序店,午餐,支出,10.00,余额\n"
	if err := os.WriteFile(csvPath, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	profile := testProfile()
	profile.Schema = "https://deg.dev/template-profile/v2"
	profile.Template.DefaultMinus = ""
	profile.Template.DefaultPlus = ""
	enabled := true
	profile.PersonalRules = []Rule{
		{Enabled: &enabled, When: `payee == "腿顺序店"`, Actions: Actions{
			From: TransferSide{Account: "Assets:Cash"},
			To:   TransferSide{Account: "Expenses:Old"},
		}},
		{Enabled: &enabled, When: `payee == "腿顺序店"`, Actions: Actions{
			PostingsMode: "append",
			Postings:     []string{"Expenses:Tip 0.50 CNY", "Assets:Cash -0.50 CNY"},
		}},
		{Enabled: &enabled, When: `payee == "腿顺序店"`, Actions: Actions{
			PostingsMode: "replace",
			Postings:     []string{"Expenses:New 10.00 CNY", "Assets:Cash -10.00 CNY"},
		}},
		{Enabled: &enabled, When: `payee == "腿顺序店"`, Actions: Actions{
			PostingsMode: "append",
			Postings:     []string{"Expenses:Fee 1.00 CNY", "Assets:Cash -1.00 CNY"},
		}},
	}
	out, err := ImportFile(profile, csvPath)
	if err != nil {
		t.Fatal(err)
	}
	assertPostingsEqual(t, postingLines(out.Orders[0]), []expectPosting{
		{Account: "Expenses:New", Amount: "10.00", Currency: "CNY"},
		{Account: "Assets:Cash", Amount: "-10.00", Currency: "CNY"},
		{Account: "Expenses:Fee", Amount: "1.00", Currency: "CNY"},
		{Account: "Assets:Cash", Amount: "-1.00", Currency: "CNY"},
	})
}
