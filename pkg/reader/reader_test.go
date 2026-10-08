package reader

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func parseConfig(t *testing.T, src string) Config {
	t.Helper()
	var cfg Config
	if err := yaml.Unmarshal([]byte(src), &cfg); err != nil {
		t.Fatalf("config: %v", err)
	}
	return cfg
}

func TestCSVKeepsBlankLinePositions(t *testing.T) {
	data := []byte("a,b\n\nc,d\n")
	plain, err := ReadBytes("x.csv", data, Config{Format: "csv"})
	if err != nil {
		t.Fatal(err)
	}
	if len(plain.Rows) != 2 {
		t.Fatalf("plain rows = %d, want 2", len(plain.Rows))
	}
	kept, err := ReadBytes("x.csv", data, Config{Format: "csv", KeepBlankLines: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(kept.Rows) != 3 || kept.Rows[1] != nil {
		t.Fatalf("kept rows = %v, want blank row at index 1", kept.Rows)
	}
}

func TestCSVDecodesGB18030AndStripsTabs(t *testing.T) {
	// "金额" in GB18030 followed by a tab-padded cell.
	gb := []byte{0xBD, 0xF0, 0xB6, 0xEE, ',', '\t', '1', '\t', '\n'}
	table, err := ReadBytes("x.csv", gb, Config{Format: "csv", Encoding: "gb18030", StripTabs: true})
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"金额", "1"}; !reflect.DeepEqual(table.Rows[0], want) {
		t.Fatalf("row = %q, want %q", table.Rows[0], want)
	}
}

func TestJSONRecordsAndColumns(t *testing.T) {
	cfg := parseConfig(t, `
format: json
records: "//result/*"
columns:
  hash: hash
  value: value
  gas: "gasUsed"
`)
	data := []byte(`{"status":"1","result":[{"hash":"0xa","value":"10","gasUsed":"21000"},{"hash":"0xb","value":"20"}]}`)
	table, err := ReadBytes("tx.json", data, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"hash", "value", "gas"}; !reflect.DeepEqual(table.Headers, want) {
		t.Fatalf("headers = %q", table.Headers)
	}
	want := [][]string{{"0xa", "10", "21000"}, {"0xb", "20", ""}}
	if !reflect.DeepEqual(table.Rows, want) {
		t.Fatalf("rows = %q, want %q", table.Rows, want)
	}
}

func TestXMLCamt053Entries(t *testing.T) {
	cfg := parseConfig(t, `
format: xml
records: "//Ntry"
columns:
  date: BookgDt/Dt
  amount: Amt
  currency: Amt/@Ccy
  direction: CdtDbtInd
  narration: NtryDtls/TxDtls/RmtInf/Ustrd
`)
	data := []byte(`<?xml version="1.0"?>
<Document xmlns="urn:iso:std:iso:20022:tech:xsd:camt.053.001.02">
 <BkToCstmrStmt><Stmt>
  <Ntry><Amt Ccy="EUR">12.50</Amt><CdtDbtInd>DBIT</CdtDbtInd><BookgDt><Dt>2026-01-02</Dt></BookgDt>
   <NtryDtls><TxDtls><RmtInf><Ustrd>Coffee</Ustrd></RmtInf></TxDtls></NtryDtls></Ntry>
  <Ntry><Amt Ccy="EUR">100.00</Amt><CdtDbtInd>CRDT</CdtDbtInd><BookgDt><Dt>2026-01-03</Dt></BookgDt></Ntry>
 </Stmt></BkToCstmrStmt>
</Document>`)
	table, err := ReadBytes("stmt.xml", data, cfg)
	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{
		{"2026-01-02", "12.50", "EUR", "DBIT", "Coffee"},
		{"2026-01-03", "100.00", "EUR", "CRDT", ""},
	}
	if !reflect.DeepEqual(table.Rows, want) {
		t.Fatalf("rows = %q, want %q", table.Rows, want)
	}
}

func TestTextRecordAndContinuation(t *testing.T) {
	cfg := parseConfig(t, `
format: text
record: '^(?P<date>\d{4}-\d{2}-\d{2})\s+(?P<narration>.+?)\s+(?P<amount>-?[\d,]+\.\d{2})\s+(?P<balance>[\d,]+\.\d{2})$'
continuation: '^\s{10,}(?P<narration>\S.*)$'
`)
	data := []byte(strings.Join([]string{
		"招商银行 对账单                     第 1 页",
		"日期        摘要                    金额        余额",
		"2026-01-02  星巴克咖啡              -32.00      1,000.00",
		"            朝阳门店",
		"2026-01-03  工资                    8,000.00    9,000.00",
		"本页合计                            7,968.00",
	}, "\n"))
	table, err := ReadBytes("stmt.txt", data, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"date", "narration", "amount", "balance"}; !reflect.DeepEqual(table.Headers, want) {
		t.Fatalf("headers = %q", table.Headers)
	}
	want := [][]string{
		{"2026-01-02", "星巴克咖啡 朝阳门店", "-32.00", "1,000.00"},
		{"2026-01-03", "工资", "8,000.00", "9,000.00"},
	}
	if !reflect.DeepEqual(table.Rows, want) {
		t.Fatalf("rows = %q, want %q", table.Rows, want)
	}
}

func TestTextRejectsPDFWithoutConverter(t *testing.T) {
	_, err := ReadBytes("stmt.pdf", []byte("%PDF-1.4"), Config{Format: "text", Record: `(?P<a>.)`})
	if err == nil || !strings.Contains(err.Error(), "convert") {
		t.Fatalf("err = %v, want converter hint", err)
	}
}

func TestCompatible(t *testing.T) {
	cases := []struct {
		cfg  Config
		file string
		want bool
	}{
		{Config{Format: "csv"}, "a.csv", true},
		{Config{Format: "csv"}, "a.txt", true},
		{Config{Format: "csv"}, "a.xlsx", false},
		{Config{Format: "xlsx"}, "a.xlsx", true},
		{Config{Format: "text"}, "a.txt", true},
		{Config{Format: "text"}, "a.pdf", false},
		{Config{Format: "text", Convert: "pdftotext-layout"}, "a.pdf", true},
		{Config{Format: "json"}, "a.json", true},
		{Config{Format: "xml"}, "a.xml", true},
	}
	for _, c := range cases {
		if got := c.cfg.Compatible(c.file); got != c.want {
			t.Errorf("%s with %s: got %v want %v", c.cfg.Format, c.file, got, c.want)
		}
	}
}

func TestXLSXExampleReads(t *testing.T) {
	path := filepath.Join("..", "..", "example", "wechat", "example-wechat-records.xlsx")
	if _, err := os.Stat(path); err != nil {
		t.Skip("example missing")
	}
	table, err := ReadFile(path, Config{Format: "xlsx"})
	if err != nil {
		t.Fatal(err)
	}
	if len(table.Rows) < 18 {
		t.Fatalf("rows = %d", len(table.Rows))
	}
	if _, err := ReadFile(path, Config{Format: "xlsx", Sheet: "nope"}); err == nil {
		t.Fatal("want sheet error")
	}
}
