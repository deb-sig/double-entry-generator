package cgb_credit

import (
	"fmt"
	"strings"
)

// parseRecord 从 CSV 行中提取广发信用卡原始字段，不做业务判断。
func (cc *CgbCredit) parseRecord(record []string) (*RawRecord, error) {
	if len(record) < 7 {
		return nil, fmt.Errorf("invalid record length: %d", len(record))
	}

	for i := range record {
		record[i] = strings.TrimSpace(record[i])
	}

	if record[0] == "" || record[0] == "交易日期" {
		return nil, nil
	}

	return &RawRecord{
		TradeDate:      record[0],
		RecordDate:     record[1],
		Description:    record[2],
		TradeAmount:    record[3],
		TradeCurrency:  record[4],
		SettleAmount:   record[5],
		SettleCurrency: record[6],
	}, nil
}
