package importer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
)

func TestSlotContractMapsBeancountFieldsAndLeavesAccountsToPersonalRules(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "bill.csv")
	body := strings.Join([]string{
		"交易时间,商户,商品,收/支,金额(元),支付方式,交易单号",
		"2026-05-21 10:30:00,一卡通充值,地铁,支出,18.90,零钱,42",
		"2026-05-22 09:00:00,公司,工资,收入,100.00,银行卡,7",
	}, "\n")
	if err := os.WriteFile(csvPath, []byte(body+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	profile := slotProfile()
	profile.TemplateRules = []Rule{{
		ID:      "不应生效的账户规则",
		Actions: Actions{From: TransferSide{Account: "Assets:ShouldNotApply"}},
	}}
	profile.PersonalRules = []Rule{{
		ID:      "一卡通",
		When:    `payee ~ "一卡通"`,
		Actions: Actions{From: TransferSide{Account: "Assets:Current:零钱"}},
	}, {
		ID:      "去掉订单号",
		When:    `metadata.orderId == "42"`,
		Actions: Actions{MetadataDrop: []string{"orderId"}},
	}}

	out, err := ImportFile(profile, csvPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Orders) != 2 {
		t.Fatalf("orders: %d", len(out.Orders))
	}
	expense := out.Orders[0]
	if expense.Peer != "一卡通充值" || expense.Item != "地铁" {
		t.Fatalf("slots: %#v", expense)
	}
	if expense.Metadata["method"] != "零钱" || expense.Metadata["type"] != "支出" {
		t.Fatalf("metadata: %#v", expense.Metadata)
	}
	if _, ok := expense.Metadata["orderId"]; ok {
		t.Fatalf("dropped orderId still present: %#v", expense.Metadata)
	}
	if joinPostings(expense) == "" || strings.Contains(joinPostings(expense), "Assets:ShouldNotApply") {
		t.Fatalf("postings: %s", joinPostings(expense))
	}
	if !strings.Contains(joinPostings(expense), "Assets:Current:零钱 -18.9") || !strings.Contains(joinPostings(expense), "CNY") {
		t.Fatalf("from: %s", joinPostings(expense))
	}
	if !strings.Contains(joinPostings(expense), "Expenses:FIXME 18.9") {
		t.Fatalf("missing expense fallback: %s", joinPostings(expense))
	}
	if !strings.EqualFold(strings.Join(expense.MetadataKeys, ","), "method,type") && strings.Join(expense.MetadataKeys, ",") != "method,type" {
		t.Fatalf("metadata order: %#v", expense.MetadataKeys)
	}

	income := out.Orders[1]
	text := joinPostings(income)
	if !strings.Contains(text, "Income:FIXME -100") || !strings.Contains(text, "Assets:FIXME 100") {
		t.Fatalf("income fallback: %s", text)
	}
}

func TestSlotContractWarnsOnRemovedColumnAndMetadataKey(t *testing.T) {
	profile := slotProfile()
	rules := []Rule{{
		ID:   "旧列",
		When: `<交易对方> ~ "一卡通" && metadata.missing == "1"`,
	}, {
		ID:   "槽位仍可用",
		When: `payee ~ "一卡通"`,
	}}
	warnings := PersonalRuleWarnings(profile, rules, "wechat@2026-09-01", "wechat@2026-04-28")
	joined := strings.Join(warnings, "\n")
	if !strings.Contains(joined, "旧列") || !strings.Contains(joined, "交易对方") || !strings.Contains(joined, "missing") {
		t.Fatalf("warnings: %s", joined)
	}
	if strings.Contains(joined, "槽位仍可用") {
		t.Fatalf("slot reference should not warn: %s", joined)
	}
	if !strings.Contains(joined, "wechat@2026-04-28") {
		t.Fatalf("version note missing: %s", joined)
	}
}

func TestSlotSkeletonDocumentsContract(t *testing.T) {
	text := SlotSkeleton("wechat@2026-04-28", slotProfile())
	for _, want := range []string{
		"template: wechat@2026-04-28",
		"#   payee: <商户>",
		"#   method: <支付方式>",
		"when: payee ~ \"商户名\"",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("skeleton missing %q:\n%s", want, text)
		}
	}
}

func slotProfile() *Profile {
	return &Profile{
		ID: "wechat",
		Template: Template{
			FileFormat:      "csv",
			DateFormat:      "yyyy-MM-dd HH:mm:ss",
			SourceHeaders:   []string{"交易时间", "商户", "商品", "收/支", "金额(元)", "支付方式", "交易单号"},
			DefaultCurrency: "CNY",
			Slots: SlotMapping{
				Date:      "<交易时间>",
				Payee:     "<商户>",
				Narration: "<商品>",
				Amount:    "<金额(元)>.number",
				Currency:  "CNY",
				Metadata: orderedStrings{
					Keys: []string{"method", "type", "orderId"},
					Values: map[string]string{
						"method":  "<支付方式>",
						"type":    "<收/支>",
						"orderId": "<交易单号>",
					},
				},
			},
			AmountSign: AmountSign{Metadata: "type", Negate: []string{"支出", "支"}},
		},
	}
}

func joinPostings(order ir.Order) string {
	parts := make([]string, 0, len(order.Postings))
	for _, posting := range order.Postings {
		parts = append(parts, posting.Line)
	}
	return strings.Join(parts, "\n")
}
