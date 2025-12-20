package bocom_credit

import (
	"fmt"
	"strings"
)

// parseRecord extracts raw fields from a CSV record.
// It only performs basic field extraction without any business logic.
func (bc *BocomCredit) parseRecord(record []string) (*RawRecord, error) {
	if len(record) < 5 {
		return nil, fmt.Errorf("invalid record length: %d", len(record))
	}

	for i := range record {
		record[i] = strings.TrimSpace(record[i])
	}

	// Skip empty rows
	if record[0] == "" {
		return nil, nil
	}

	return &RawRecord{
		TradeDate:            record[0],
		RecordDate:           record[1],
		TradeDescription:     record[2],
		TxnCurrencyAmount:    record[3],
		SettleCurrencyAmount: record[4],
	}, nil
}
