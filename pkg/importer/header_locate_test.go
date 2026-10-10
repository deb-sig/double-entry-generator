package importer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func headerLocateProfile() *Profile {
	return &Profile{
		ID:                   "header-locate",
		RequiredCapabilities: []string{"template.headerLocate"},
		Template: Template{
			FileFormat:        "csv",
			DateFormat:        "yyyy-MM-dd",
			HeaderLocate:      true,
			HeaderScanMaxRows: 40,
			SourceHeaders:     []string{"日期", "金额", "商家"},
			DefaultMinus:      "Assets:FIXME",
			DefaultPlus:       "Expenses:FIXME",
			DefaultCurrency:   "CNY",
			Columns: ColumnMapping{
				Date:   "日期",
				Amount: "金额",
				Payee:  "商家",
			},
		},
	}
}

func writeCSV(t *testing.T, name, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestHeaderLocateShiftedPreamble(t *testing.T) {
	csvPath := writeCSV(t, "shifted.csv",
		"账单导出说明\n导出时间: 2026-01-01\n日期,金额,商家\n2026-05-21,18.90,午餐\n")
	profile := headerLocateProfile()
	out, err := ImportFile(profile, csvPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Orders) != 1 {
		t.Fatalf("expected 1 order, got %d", len(out.Orders))
	}
	if out.Orders[0].Peer != "午餐" {
		t.Fatalf("payee=%q", out.Orders[0].Peer)
	}
}

func TestHeaderLocateColumnReorder(t *testing.T) {
	csvPath := writeCSV(t, "reorder.csv",
		"商家,日期,金额\n午餐,2026-05-21,18.90\n")
	profile := headerLocateProfile()
	out, err := ImportFile(profile, csvPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Orders) != 1 {
		t.Fatalf("expected 1 order, got %d", len(out.Orders))
	}
	if out.Orders[0].Peer != "午餐" {
		t.Fatalf("payee=%q want 午餐 (name remap after reorder)", out.Orders[0].Peer)
	}
	if out.Orders[0].Money < 18.89 || out.Orders[0].Money > 18.91 {
		t.Fatalf("money=%v want ~18.90", out.Orders[0].Money)
	}
}

func TestHeaderLocateDuplicateColumnsError(t *testing.T) {
	csvPath := writeCSV(t, "dup.csv",
		"日期,金额,金额,商家\n2026-05-21,18.90,1.00,午餐\n")
	profile := headerLocateProfile()
	_, err := ImportFile(profile, csvPath)
	if err == nil {
		t.Fatal("expected duplicate column error")
	}
	if !strings.Contains(err.Error(), "duplicate column name") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHeaderLocateAmbiguousTwoHeaderRows(t *testing.T) {
	csvPath := writeCSV(t, "ambig.csv",
		"日期,金额,商家\n摘要行,0,占位\n日期,金额,商家\n2026-05-21,18.90,午餐\n")
	profile := headerLocateProfile()
	_, err := ImportFile(profile, csvPath)
	if err == nil {
		t.Fatal("expected ambiguous header error")
	}
	if !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHeaderLocateMissingField(t *testing.T) {
	csvPath := writeCSV(t, "missing.csv",
		"日期,商家\n2026-05-21,午餐\n")
	profile := headerLocateProfile()
	_, err := ImportFile(profile, csvPath)
	if err == nil {
		t.Fatal("expected missing header error")
	}
	if !strings.Contains(err.Error(), "headerLocate") && !strings.Contains(err.Error(), "not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHeaderLocateHeaderlessPathUnchanged(t *testing.T) {
	csvPath := writeCSV(t, "headerless.csv",
		"2026-05-21,10:30:00,午餐,,18.90\n2026-05-22,09:00:00,工资,100.00,\n")
	profile := &Profile{
		ID: "headerless",
		Template: Template{
			FileFormat:      "csv",
			DateFormat:      "yyyy-MM-dd HH:mm:ss",
			HeaderLocate:    false,
			SourceHeaders:   []string{"date", "time", "payee", "in", "out"},
			DefaultMinus:    "Assets:FIXME",
			DefaultPlus:     "Expenses:FIXME",
			DefaultCurrency: "CNY",
			Columns: ColumnMapping{
				Date:      "date",
				Time:      "time",
				AmountIn:  "in",
				AmountOut: "out",
				Payee:     "payee",
			},
		},
	}
	out, err := ImportFile(profile, csvPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Orders) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(out.Orders))
	}
}

func TestHeaderLocateCapabilityGate(t *testing.T) {
	if _, ok := SupportedCapabilities["template.headerLocate"]; !ok {
		t.Fatal("template.headerLocate missing from SupportedCapabilities")
	}
	profile := &Profile{
		RequiredCapabilities: []string{"template.headerLocate"},
	}
	if err := profile.ValidateCapabilities(); err != nil {
		t.Fatalf("expected capability accepted: %v", err)
	}
	profile.RequiredCapabilities = []string{"template.headerLocate", "actions.teleport"}
	if err := profile.ValidateCapabilities(); err == nil {
		t.Fatal("expected unknown capability to fail closed")
	}
}
