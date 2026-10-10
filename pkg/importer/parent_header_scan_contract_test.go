package importer

import (
	"strings"
	"testing"
)

func TestParentHeaderLocateEmptyHeadersFailClosed(t *testing.T) {
	for _, headers := range [][]string{nil, {"", " "}} {
		p := headerLocateProfile()
		p.Template.SourceHeaders = headers
		_, err := ImportFile(p, writeCSV(t, "empty.csv", "日期,金额,商家\n2026-05-21,18.90,午餐\n"))
		if err == nil || !strings.Contains(err.Error(), "sourceHeaders") {
			t.Fatalf("headers=%v: %v", headers, err)
		}
	}
}

func TestParentHeaderLocateZeroUnlimitedAndPositiveBoundary(t *testing.T) {
	p := headerLocateProfile()
	path := writeCSV(t, "late.csv", strings.Repeat("preamble\n", 70)+"日期,金额,商家\n2026-05-21,18.90,午餐\n")
	for _, limit := range []int{0, 70, 71} {
		p.Template.HeaderScanMaxRows = limit
		out, err := ImportFile(p, path)
		if limit == 70 {
			if err == nil {
				t.Fatal("70 rows must not reach header at index 70")
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		if len(out.Orders) != 1 || out.Orders[0].Peer != "午餐" {
			t.Fatalf("limit=%d: %+v", limit, out)
		}
	}
}
