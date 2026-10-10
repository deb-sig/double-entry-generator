package importer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
	"gopkg.in/yaml.v3"
)

func TestWhenOperatorsStartsEndsRegex(t *testing.T) {
	row := Row{
		Payee:     "美团外卖",
		Narration: "工作日午餐",
		Amount:    "32.50",
		Raw:       map[string]string{"商家": "美团外卖", "说明": "工作日午餐"},
	}
	order := ir.Order{}
	cases := []struct {
		when string
		want bool
	}{
		{`payee ^= "美团"`, true},
		{`payee ^= "饿了么"`, false},
		{`payee $= "外卖"`, true},
		{`narration =~ ".*午餐.*"`, true},
		{`narration =~ "午餐"`, false}, // not full-string match
		{`payee ~ "团" && narration !~ "晚餐"`, true},
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

func TestReplaceColumnMethod(t *testing.T) {
	row := Row{Raw: map[string]string{"商品说明": "咖啡店午餐"}}
	got := renderRuleText(`<商品说明>.replace("咖啡","早餐")`, row, ir.Order{})
	if got != "早餐店午餐" {
		t.Fatalf("replace got %q", got)
	}
}

func TestQuotedActionLiteralsAreNotEvaluated(t *testing.T) {
	row := Row{
		Payee: "真实商家",
		Raw:   map[string]string{"金额": "99.00", "商家": "真实商家"},
	}
	order := ir.Order{}
	cases := []struct {
		in   string
		want string
	}{
		{`"<金额>"`, "<金额>"},
		{`"payee"`, "payee"},
		{`"1+2"`, "1+2"},
		{`" 前后空格 "`, " 前后空格 "},
		{`"含\"引号\"\\路径\n下一行"`, "含\"引号\"\\路径\n下一行"},
		{`<商家>`, "真实商家"},
	}
	for _, tc := range cases {
		got := resolveActionValue(tc.in, row, order)
		if got != tc.want {
			t.Fatalf("resolveActionValue(%q)=%q want %q", tc.in, got, tc.want)
		}
	}
	if got := renderPostingText(`"1+2"`, row, order); got != "1+2" {
		t.Fatalf("literal amount must skip arithmetic, got %q", got)
	}
}

func TestDynamicTagAndLinkFromColumn(t *testing.T) {
	row := Row{Raw: map[string]string{"标签列": "food", "链接列": "meal-1"}}
	order := ir.Order{}
	if got := resolveActionValue(`<标签列>`, row, order); got != "food" {
		t.Fatalf("tag ref=%q", got)
	}
	if got := resolveActionValue(`<链接列>`, row, order); got != "meal-1" {
		t.Fatalf("link ref=%q", got)
	}
}

func TestRequiredCapabilitiesRejectUnknown(t *testing.T) {
	profile := &Profile{
		RequiredCapabilities: []string{"actions.teleport"},
		ProtocolVersion:      "mirato-deg-rules/1",
	}
	if err := profile.ValidateCapabilities(); err == nil {
		t.Fatal("expected unsupported capability error")
	}
	profile.RequiredCapabilities = []string{"actions.quoted_literal", "rule.templateId"}
	if err := profile.ValidateCapabilities(); err != nil {
		t.Fatal(err)
	}
}

func TestTemplateIDScopePreventsGlobalMisuse(t *testing.T) {
	enabled := true
	profile := &Profile{
		ID:     "alipay",
		Schema: "https://double-entry-generator/schema/v2",
		Template: Template{
			DefaultCurrency: "CNY",
			DefaultMinus:    "Assets:Cash",
			DefaultPlus:     "Expenses:FIXME",
			Columns:         ColumnMapping{Date: "日期", Amount: "金额", Payee: "商家"},
			SourceHeaders:   []string{"日期", "金额", "商家"},
		},
		PersonalRules: []Rule{
			{
				ID:         "scoped",
				TemplateID: "wechat",
				When:       `payee ~ "咖啡"`,
				Enabled:    &enabled,
				Actions: Actions{
					To: TransferSide{Account: "Expenses:Food"},
				},
			},
			{
				ID:      "global",
				When:    `payee ~ "咖啡"`,
				Enabled: &enabled,
				Actions: Actions{
					To:       TransferSide{Account: "Expenses:Drink"},
					From:     TransferSide{Account: "Assets:Cash"},
					Date:     "<日期>",
					Amount:   "<金额>",
					Currency: "CNY",
				},
			},
		},
	}
	dir := t.TempDir()
	bill := filepath.Join(dir, "bill.csv")
	if err := os.WriteFile(bill, []byte("日期,金额,商家\n2026-09-01,12.00,咖啡店\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	irData, err := ImportFile(profile, bill)
	if err != nil {
		t.Fatal(err)
	}
	if len(irData.Orders) != 1 {
		t.Fatalf("orders=%d", len(irData.Orders))
	}
	// scoped wechat rule must not apply on alipay profile
	foundFood := false
	for _, p := range irData.Orders[0].Postings {
		if strings.Contains(p.Line, "Expenses:Food") {
			foundFood = true
		}
	}
	if foundFood {
		t.Fatalf("scoped templateId=wechat rule applied on alipay: %#v", irData.Orders[0].Postings)
	}
}

func TestMiratoRustExportedYAMLFile(t *testing.T) {
	// Real cross-engine: YAML written by Mirato `DEG_INTEROP_OUT` export, stored under testdata/.
	raw, err := os.ReadFile(filepath.Join("testdata", "mirato-exported-personal-rules.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		ProtocolVersion      string   `yaml:"protocolVersion"`
		RequiredCapabilities []string `yaml:"requiredCapabilities"`
		PersonalRules        []Rule   `yaml:"personalRules"`
	}
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.ProtocolVersion == "" || len(doc.RequiredCapabilities) == 0 {
		t.Fatalf("exported file missing protocol annotations: %#v", doc)
	}
	falseVal := false
	for i := range doc.PersonalRules {
		if doc.PersonalRules[i].ID == "disabled" {
			doc.PersonalRules[i].Enabled = &falseVal
		}
	}
	profile := &Profile{
		ID:                   "alipay",
		Schema:               "https://double-entry-generator/schema/v2",
		ProtocolVersion:      doc.ProtocolVersion,
		RequiredCapabilities: doc.RequiredCapabilities,
		Template: Template{
			DefaultCurrency: "CNY",
			SkipLeadingRows: 0,
			Columns: ColumnMapping{
				Date: "日期", Amount: "金额", Payee: "商家", Narration: "说明",
			},
			SourceHeaders: []string{"日期", "金额", "商家", "说明", "交易状态"},
		},
		PersonalRules: append([]Rule{{
			ID: "base",
			Actions: Actions{
				Date: "<日期>", Amount: "<金额>.number", Currency: "CNY",
				Payee: "<商家>", Narration: "<说明>",
				From: TransferSide{Account: "Assets:FIXME"},
				To:   TransferSide{Account: "Expenses:FIXME"},
			},
		}}, doc.PersonalRules...),
	}
	dir := t.TempDir()
	bill := filepath.Join(dir, "bill.csv")
	csv := strings.Join([]string{
		"日期,金额,商家,说明,交易状态",
		"2026-09-01,12.30,咖啡店,工作日午餐,交易成功",
		"2026-09-03,1.00,关闭店,x,交易关闭",
		"2026-09-02,3.00,美团骑手,配送,交易成功",
	}, "\n") + "\n"
	if err := os.WriteFile(bill, []byte(csv), 0o644); err != nil {
		t.Fatal(err)
	}
	irData, err := ImportFile(profile, bill)
	if err != nil {
		t.Fatal(err)
	}
	if len(irData.Orders) != 2 {
		t.Fatalf("want 2 orders after ignore, got %d", len(irData.Orders))
	}
	if irData.Orders[0].Flag != "!" || irData.Orders[0].Metadata["source"] != "mirato" {
		t.Fatalf("coffee order=%#v", irData.Orders[0])
	}
	hasDelivery := false
	for _, p := range irData.Orders[1].Postings {
		if strings.Contains(p.Line, "Expenses:Delivery") {
			hasDelivery = true
		}
	}
	if !hasDelivery {
		t.Fatalf("starts rule missing: %#v", irData.Orders[1].Postings)
	}
}

func TestMiratoExportedRulesFixture(t *testing.T) {
	// Fixture produced by Mirato deg_rules export (kept in-repo for cross-engine CI).
	yamlText := `
personalRules:
  - id: ignore-closed
    name: 忽略关闭
    when: <交易状态> == "交易关闭"
    actions:
      ignore: true
  - id: coffee
    name: 咖啡
    when: (payee ~ "咖啡") && (narration ~ "午餐")
    actions:
      from: Assets:Cash
      to: Expenses:Food
      flag: "!"
      link: meal-link
      tag: food
      metadata:
        source: mirato
  - id: disabled
    name: 禁用
    enabled: false
    when: payee ~ "禁用商户"
    actions:
      to: Expenses:ShouldNotApply
  - id: scoped-wechat
    name: 仅微信
    templateId: wechat
    when: payee ~ "咖啡"
    actions:
      to: Expenses:WrongScope
  - id: starts
    name: 前缀
    when: payee ^= "美团"
    actions:
      to: Expenses:Delivery
      from: Assets:Cash
`
	var doc struct {
		PersonalRules []Rule `yaml:"personalRules"`
	}
	if err := yaml.Unmarshal([]byte(yamlText), &doc); err != nil {
		t.Fatal(err)
	}
	falseVal := false
	for i := range doc.PersonalRules {
		if doc.PersonalRules[i].ID == "disabled" {
			doc.PersonalRules[i].Enabled = &falseVal
		}
	}
	profile := &Profile{
		ID:     "alipay",
		Schema: "https://double-entry-generator/schema/v2",
		Template: Template{
			DefaultCurrency: "CNY",
			SkipLeadingRows: 0,
			Columns: ColumnMapping{
				Date:      "日期",
				Amount:    "金额",
				Payee:     "商家",
				Narration: "说明",
			},
			SourceHeaders: []string{"日期", "金额", "商家", "说明", "交易状态"},
			Metadata:      map[string]string{},
		},
		PersonalRules: append([]Rule{{
			ID: "base",
			Actions: Actions{
				Date:      "<日期>",
				Amount:    "<金额>.number",
				Currency:  "CNY",
				Payee:     "<商家>",
				Narration: "<说明>",
				From:      TransferSide{Account: "Assets:FIXME"},
				To:        TransferSide{Account: "Expenses:FIXME"},
			},
		}}, doc.PersonalRules...),
	}
	dir := t.TempDir()
	bill := filepath.Join(dir, "bill.csv")
	csv := strings.Join([]string{
		"日期,金额,商家,说明,交易状态",
		"2026-09-01,12.30,咖啡店,工作日午餐,交易成功",
		"2026-09-01,-8.00,咖啡店,退款,交易成功",
		"2026-09-02,3.00,美团骑手,配送,交易成功",
		"2026-09-03,1.00,关闭店,x,交易关闭",
		"2026-09-04,9.00,禁用商户,x,交易成功",
	}, "\n") + "\n"
	if err := os.WriteFile(bill, []byte(csv), 0o644); err != nil {
		t.Fatal(err)
	}
	irData, err := ImportFile(profile, bill)
	if err != nil {
		t.Fatal(err)
	}
	// closed ignored; 4 remaining (incl negative amount row)
	if len(irData.Orders) != 4 {
		t.Fatalf("want 4 orders after ignore, got %d", len(irData.Orders))
	}
	coffee := irData.Orders[0]
	if coffee.Flag != "!" {
		t.Fatalf("flag=%q", coffee.Flag)
	}
	if len(coffee.Links) != 1 || coffee.Links[0] != "meal-link" {
		t.Fatalf("links=%v", coffee.Links)
	}
	if coffee.Metadata["source"] != "mirato" {
		t.Fatalf("metadata=%v", coffee.Metadata)
	}
	hasFood := false
	for _, p := range coffee.Postings {
		if strings.Contains(p.Line, "Expenses:Food") {
			hasFood = true
		}
		if strings.Contains(p.Line, "Expenses:WrongScope") {
			t.Fatalf("scoped rule leaked: %v", coffee.Postings)
		}
	}
	if !hasFood {
		t.Fatalf("coffee not classified: %#v", coffee.Postings)
	}
	// disabled merchant should stay FIXME, not Expenses:ShouldNotApply
	for _, o := range irData.Orders {
		if o.Peer == "禁用商户" {
			for _, p := range o.Postings {
				if strings.Contains(p.Line, "ShouldNotApply") {
					t.Fatalf("disabled rule applied: %v", o.Postings)
				}
			}
		}
		if o.Peer == "美团骑手" {
			ok := false
			for _, p := range o.Postings {
				if strings.Contains(p.Line, "Expenses:Delivery") {
					ok = true
				}
			}
			if !ok {
				t.Fatalf("starts_with not applied: %#v", o.Postings)
			}
		}
	}
	// Negative bill amount: compare posting signs, not only account names.
	// This row does not match the coffee when (narration≠午餐), so accounts stay
	// defaults — still assert the signed cash leg exists.
	refund := irData.Orders[1]
	if refund.Money != 8.00 {
		t.Fatalf("refund abs money=%v (float64 short-decimal only)", refund.Money)
	}
	sawMinusCash := false
	for _, p := range refund.Postings {
		if strings.Contains(p.Line, "Assets:") && strings.Contains(p.Line, "-") {
			sawMinusCash = true
		}
	}
	if !sawMinusCash {
		t.Fatalf("refund postings missing signed asset leg: %#v", refund.Postings)
	}
}
