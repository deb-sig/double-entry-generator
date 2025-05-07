package jd

import (
	"testing"
)

func TestTranslateValue(t *testing.T) {
	cases := []struct {
		str    string
		res    int64
		resErr bool
	}{
		{"131.77(已退款89.84)", 13177, false},
		{"468.32(已全额退款)", 46832, false},
		{"397.84", 39784, false},
		{"233.33(已全额退款)333", 0, true},
		{"233.1(已全额)", 0, true},
		{"223.1(已退款)", 0, true},
		{"233.11(已退款11.11)22", 0, true},
	}
	jd := &JD{}

	for _, c := range cases {
		got, err := jd.translateValue(c.str)
		if (err != nil) != c.resErr {
			t.Errorf("translateValue() error = %v, wantErr %v", err, c.resErr)
			return
		}
		if got != c.res {
			t.Errorf("translateValue() = %v, want %v", got, c.res)
		}
	}
}
