package xls

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"time"

	yymmdd "github.com/extrame/goyymmdd"
)

// content type
type contentHandler interface {
	String(*WorkBook) []string
	FirstCol() uint16
	LastCol() uint16
}

type Col struct {
	RowB      uint16
	FirstColB uint16
}

type Coler interface {
	Row() uint16
}

func (c *Col) Row() uint16 {
	return c.RowB
}

func (c *Col) FirstCol() uint16 {
	return c.FirstColB
}

func (c *Col) LastCol() uint16 {
	return c.FirstColB
}

func (c *Col) String(wb *WorkBook) []string {
	return []string{"default"}
}

type XfRk struct {
	Index uint16
	Rk    RK
}

func (xf *XfRk) String(wb *WorkBook) string {
	idx := int(xf.Index)
	if len(wb.Xfs) > idx {
		fNo := wb.Xfs[idx].formatNo()
		if fNo >= 164 { // user defined format
			if formatter := wb.Formats[fNo]; formatter != nil {
				formatterLower := strings.ToLower(formatter.str)
				if formatterLower == "general" ||
					strings.Contains(formatter.str, "#") ||
					strings.Contains(formatter.str, ".00") ||
					strings.Contains(formatterLower, "m/y") ||
					strings.Contains(formatterLower, "d/y") ||
					strings.Contains(formatterLower, "m.y") ||
					strings.Contains(formatterLower, "d.y") ||
					strings.Contains(formatterLower, "h:") ||
					strings.Contains(formatterLower, "д.г") {
					//If format contains # or .00 then this is a number
					return xf.Rk.String()
				} else {
					i, f, isFloat := xf.Rk.number()
					if !isFloat {
						f = float64(i)
					}
					t := timeFromExcelTime(f, wb.dateMode == 1)
					return yymmdd.Format(t, formatter.str)
				}
			}
			// see http://www.openoffice.org/sc/excelfileformat.pdf Page #174
		} else if 14 <= fNo && fNo <= 17 || fNo == 22 || 27 <= fNo && fNo <= 36 || 50 <= fNo && fNo <= 58 { // jp. date format
			i, f, isFloat := xf.Rk.number()
			if !isFloat {
				f = float64(i)
			}
			t := timeFromExcelTime(f, wb.dateMode == 1)
			return t.Format(time.RFC3339) //TODO it should be international
		}
	}
	return xf.Rk.String()
}

type RK uint32

func (rk RK) number() (intNum int64, floatNum float64, isFloat bool) {
	multiplied := rk & 1
	isInt := rk & 2
	val := int32(rk) >> 2
	if isInt == 0 {
		isFloat = true
		floatNum = math.Float64frombits(uint64(val) << 34)
		if multiplied != 0 {
			floatNum = floatNum / 100
		}
		return
	}
	if multiplied != 0 {
		isFloat = true
		floatNum = float64(val) / 100
		return
	}
	return int64(val), 0, false
}

func (rk RK) String() string {
	i, f, isFloat := rk.number()
	if isFloat {
		return strconv.FormatFloat(f, 'f', -1, 64)
	}
	return strconv.FormatInt(i, 10)
}

var ErrIsInt = fmt.Errorf("is int")

func (rk RK) Float() (float64, error) {
	_, f, isFloat := rk.number()
	if !isFloat {
		return 0, ErrIsInt
	}
	return f, nil
}

type MulrkCol struct {
	Col
	Xfrks    []XfRk
	LastColB uint16
}

func (c *MulrkCol) LastCol() uint16 {
	return c.LastColB
}

func (c *MulrkCol) String(wb *WorkBook) []string {
	var res = make([]string, len(c.Xfrks))
	for i := 0; i < len(c.Xfrks); i++ {
		xfrk := c.Xfrks[i]
		res[i] = xfrk.String(wb)
	}
	return res
}

type MulBlankCol struct {
	Col
	Xfs      []uint16
	LastColB uint16
}

func (c *MulBlankCol) LastCol() uint16 {
	return c.LastColB
}

func (c *MulBlankCol) String(wb *WorkBook) []string {
	return make([]string, len(c.Xfs))
}

type NumberCol struct {
	Col
	Index uint16
	Float float64
}

func (c *NumberCol) String(wb *WorkBook) []string {
	if int(c.Index) < len(wb.Xfs) {
		fNo := wb.Xfs[c.Index].formatNo()
		if isBuiltinDateFormat(fNo) {
			return []string{timeFromExcelTime(c.Float, wb.dateMode == 1).Format(time.RFC3339)}
		}
		if f := wb.Formats[fNo]; fNo >= 164 && f != nil && isDateFormatString(f.str) {
			return []string{yymmdd.Format(timeFromExcelTime(c.Float, wb.dateMode == 1), f.str)}
		}
	}
	return []string{strconv.FormatFloat(c.Float, 'f', -1, 64)}
}

// isBuiltinDateFormat reports the built-in date/time format ids
// (openoffice excelfileformat.pdf p.174), the same set XfRk uses.
func isBuiltinDateFormat(fNo uint16) bool {
	return 14 <= fNo && fNo <= 17 || fNo == 22 || 27 <= fNo && fNo <= 36 || 50 <= fNo && fNo <= 58
}

// isDateFormatString is true for user formats that render dates or times
// rather than numbers: they mention a date/time token and no digit
// placeholder.
func isDateFormatString(format string) bool {
	f := strings.ToLower(format)
	if f == "general" || strings.ContainsAny(f, "#0?") {
		return false
	}
	return strings.ContainsAny(f, "ymdhs")
}

type FormulaStringCol struct {
	Col
	RenderedValue string
}

func (c *FormulaStringCol) String(wb *WorkBook) []string {
	return []string{c.RenderedValue}
}

//str, err = wb.get_string(buf_item, size)
//wb.sst[offset_pre] = wb.sst[offset_pre] + str

type FormulaCol struct {
	Header struct {
		Col
		IndexXf uint16
		Result  [8]byte
		Flags   uint16
		_       uint32
	}
	Bts []byte
}

func (c *FormulaCol) String(wb *WorkBook) []string {
	return []string{"FormulaCol"}
}

type RkCol struct {
	Col
	Xfrk XfRk
}

func (c *RkCol) String(wb *WorkBook) []string {
	return []string{c.Xfrk.String(wb)}
}

type LabelsstCol struct {
	Col
	Xf  uint16
	Sst uint32
}

func (c *LabelsstCol) String(wb *WorkBook) []string {
	return []string{wb.sst[int(c.Sst)]}
}

type labelCol struct {
	BlankCol
	Str string
}

func (c *labelCol) String(wb *WorkBook) []string {
	return []string{c.Str}
}

type BlankCol struct {
	Col
	Xf uint16
}

func (c *BlankCol) String(wb *WorkBook) []string {
	return []string{""}
}
