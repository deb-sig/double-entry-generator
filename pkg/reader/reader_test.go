package reader

import (
	"bytes"
	"encoding/base64"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"
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

func TestHTMLTableRows(t *testing.T) {
	cfg := parseConfig(t, `
format: html
records: "//table[@id='txns']//tr[position()>1]"
columns:
  date: td[1]
  narration: td[2]
  amount: td[3]
`)
	data := []byte(`<html><body><table id="txns">
<tr><th>日期</th><th>摘要</th><th>金额</th></tr>
<tr><td>2026-01-02</td><td> 星巴克 <b>朝阳</b> </td><td>-32.00</td></tr>
<tr><td>2026-01-03</td><td>工资</td><td>8,000.00</td></tr>
</table></body></html>`)
	table, err := ReadBytes("mail.html", data, cfg)
	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{{"2026-01-02", "星巴克 朝阳", "-32.00"}, {"2026-01-03", "工资", "8,000.00"}}
	if !reflect.DeepEqual(table.Rows, want) {
		t.Fatalf("rows = %q", table.Rows)
	}
}

func TestEMLPicksHTMLPartAndAttachment(t *testing.T) {
	html := "<html><body><table id=\"t\"><tr><td>2026-01-02</td><td>咖啡</td><td>-5.00</td></tr></table></body></html>"
	csv := "d,a\n2026-01-05,1.50\n"
	eml := strings.Join([]string{
		"From: bank@example.com",
		"Subject: =?UTF-8?B?6LSm5Y2V?=",
		"MIME-Version: 1.0",
		"Content-Type: multipart/mixed; boundary=\"outer\"",
		"",
		"--outer",
		"Content-Type: multipart/alternative; boundary=\"inner\"",
		"",
		"--inner",
		"Content-Type: text/plain; charset=utf-8",
		"",
		"plain text",
		"--inner",
		"Content-Type: text/html; charset=utf-8",
		"Content-Transfer-Encoding: base64",
		"",
		base64Lines(html),
		"--inner--",
		"--outer",
		"Content-Type: text/csv; name=\"bill.csv\"",
		"Content-Disposition: attachment; filename=\"bill.csv\"",
		"Content-Transfer-Encoding: quoted-printable",
		"",
		"d,a\n2026-01-05,1.50\n",
		"--outer--",
		"",
	}, "\r\n")

	htmlCfg := parseConfig(t, `
format: eml
part: text/html
inner:
  format: html
  records: "//table[@id='t']//tr"
  columns: { date: "td[1]", narration: "td[2]", amount: "td[3]" }
`)
	table, err := ReadBytes("bill.eml", []byte(eml), htmlCfg)
	if err != nil {
		t.Fatal(err)
	}
	if want := [][]string{{"2026-01-02", "咖啡", "-5.00"}}; !reflect.DeepEqual(table.Rows, want) {
		t.Fatalf("html part rows = %q", table.Rows)
	}

	csvCfg := parseConfig(t, `
format: eml
part: "attachment:*.csv"
inner: { format: csv }
`)
	table, err = ReadBytes("bill.eml", []byte(eml), csvCfg)
	if err != nil {
		t.Fatal(err)
	}
	if want := [][]string{{"d", "a"}, {"2026-01-05", "1.50"}}; !reflect.DeepEqual(table.Rows, want) {
		t.Fatalf("attachment rows = %q (csv=%q)", table.Rows, csv)
	}

	_, err = ReadBytes("bill.eml", []byte(eml), parseConfig(t, "format: eml\npart: attachment:*.pdf\ninner: {format: text, record: '(?P<a>.)'}"))
	if err == nil || !strings.Contains(err.Error(), "no part matches") {
		t.Fatalf("err = %v", err)
	}
}

func base64Lines(s string) string {
	enc := base64.StdEncoding.EncodeToString([]byte(s))
	var b strings.Builder
	for len(enc) > 76 {
		b.WriteString(enc[:76] + "\r\n")
		enc = enc[76:]
	}
	b.WriteString(enc)
	return b.String()
}

func TestAPIExpandURLAndResponse(t *testing.T) {
	t.Setenv("DEG_TEST_KEY", "k3y")
	got, err := ExpandURL("https://x.test/api?address={source}&apikey={env.DEG_TEST_KEY}", "0xAB C")
	if err != nil || got != "https://x.test/api?address=0xAB+C&apikey=k3y" {
		t.Fatalf("got %q err %v", got, err)
	}
	if _, err := ExpandURL("https://x.test/?k={env.DEG_MISSING_KEY}", "s"); err == nil || !strings.Contains(err.Error(), "DEG_MISSING_KEY") {
		t.Fatalf("missing env should error, got %v", err)
	}
	cfg := parseConfig(t, `
format: api
url: "https://x.test/api?address={source}"
records: "//result/*"
columns: { hash: hash, value: value }
`)
	if !cfg.Compatible("0xabc") {
		t.Fatal("api reader should accept any source")
	}
	table, err := ReadBytes("0xabc", []byte(`{"result":[{"hash":"0x1","value":"2"}]}`), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if want := [][]string{{"0x1", "2"}}; !reflect.DeepEqual(table.Rows, want) {
		t.Fatalf("rows = %q", table.Rows)
	}
}

// excelize applies custom number formats: a stored 0.85 in a "0.000" cell
// reads as "0.850". Amounts must keep the stored value's digits; dates keep
// their shown text.
func TestXLSXNumberFormatDoesNotPadAmounts(t *testing.T) {
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()
	sheet := f.GetSheetName(0)
	three, err := f.NewStyle(&excelize.Style{CustomNumFmt: strPtr("0.000")})
	if err != nil {
		t.Fatal(err)
	}
	date, err := f.NewStyle(&excelize.Style{CustomNumFmt: strPtr("yyyy-mm-dd")})
	if err != nil {
		t.Fatal(err)
	}
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(f.SetSheetRow(sheet, "A1", &[]any{"日期", "金额", "账号"}))
	must(f.SetCellValue(sheet, "A2", time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)))
	must(f.SetCellStyle(sheet, "A2", "A2", date))
	must(f.SetCellValue(sheet, "B2", 0.85))
	must(f.SetCellStyle(sheet, "B2", "B2", three))
	must(f.SetCellValue(sheet, "C2", "001.100"))
	var buf bytes.Buffer
	must(f.Write(&buf))

	table, err := readXLSX(buf.Bytes(), Config{})
	if err != nil {
		t.Fatal(err)
	}
	got := table.Rows[1]
	if got[0] != "2024-01-02" || got[1] != "0.85" || got[2] != "001.100" {
		t.Fatalf("row = %q, want date shown, amount unpadded, text untouched", got)
	}
}

func strPtr(s string) *string { return &s }
