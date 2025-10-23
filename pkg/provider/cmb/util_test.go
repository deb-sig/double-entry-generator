package cmb

import "testing"

func TestExtractCreditAmount(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  float64
	}{
		{name: "simple", input: "1234.56", want: 1234.56},
		{name: "negative", input: "-1234.56", want: 1234.56},
		{name: "thousands", input: "-1,234.56", want: 1234.56},
		{name: "integer thousands", input: "1,234", want: 1234},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := extractCreditAmount(tc.input)
			if err != nil {
				t.Fatalf("extractCreditAmount(%q) returned error: %v", tc.input, err)
			}
			if got != tc.want {
				t.Fatalf("extractCreditAmount(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestExtractCreditAmountInvalid(t *testing.T) {
	if _, err := extractCreditAmount("invalid"); err == nil {
		t.Fatalf("expected error for invalid input")
	}
}
