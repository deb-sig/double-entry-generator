package hxsec

import (
	"bufio"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/deb-sig/double-entry-generator/v2/pkg/io/reader"
	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
)

const (
	colSettlementDate   = 0  // 交收日期
	colTradeTime        = 1  // 成交时间
	colBusinessType     = 2  // 业务名称
	colSecurityCode     = 3  // 证券代码
	colSecurityName     = 4  // 证券名称
	colTradePrice       = 5  // 成交价格
	colTradeVolume      = 6  // 成交数量
	colRemainingVolume  = 7  // 剩余数量
	colTradeAmount      = 8  // 成交金额
	colTotalFee         = 9  // 总费用
	colSettlementAmount = 10 // 清算金额
	colCurrentBalance   = 11 // 资金本次余额
	colNetCommission    = 12 // 净佣金
	colStampTax         = 13 // 印花税
	colTransferFee      = 14 // 过户费
	colSettlementFee    = 15 // 清算费
	colRegulatoryFee    = 16 // 交易规费
	colHandlingFee      = 17 // 经手费
	colSecuritiesFee    = 18 // 证管费
	colFinancialLevy    = 19 // 财汇局征费
	colOtherFees        = 20 // 其他费
	colCurrency         = 21 // 币种
	colTradeID          = 22 // 成交编号
	colShareholderCode  = 23 // 股东代码
	colAccountNumber    = 24 // 资金帐号
	LocalTimeFmt        = "2006-01-02 15:04:05 -0700"
)

type Statistics struct {
	LineNum int
}

type Hxsec struct {
	Statistics Statistics `json:"statistics,omitempty"`
	LineNum    int        `json:"line_num,omitempty"`
	ir         *ir.IR
}

func New() *Hxsec {
	return &Hxsec{
		Statistics: Statistics{},
		LineNum:    0,
		ir:         ir.New(),
	}
}

func (h *Hxsec) fieldsToIR(fields []string) error {
	// trim strings
	for idx, a := range fields {
		a = strings.Trim(a, " ")
		a = strings.Trim(a, "\t")
		fields[idx] = a
	}

	// Create IR order based directly on transaction type
	irO := ir.Order{
		// OrderType will be set specifically below based on TypeOriginal
		Peer:         "hxsec",
		TypeOriginal: fields[colBusinessType],
		Metadata:     map[string]string{},
	}

	// set common metadata
	switch fields[colBusinessType] {
	case "证券买入", "融券回购", "证券卖出", "融券购回", "红利入账":
		if len(fields[colTradeID]) > 0 {
			// "红利入账" have no trade_id
			irO.Metadata["trade_id"] = fields[colTradeID]
		}
		irO.Metadata["security_code"] = fields[colSecurityCode]
		irO.Metadata["security_name"] = fields[colSecurityName]
		irO.Metadata["shareholder_code"] = fields[colShareholderCode]
		irO.Metadata["account_number"] = fields[colAccountNumber]
	}

	switch fields[colBusinessType] {
	case "银行转证券":
		irO.OrderType = ir.OrderTypeChinaSecuritiesBankTransferToBroker
		amount, err := strconv.ParseFloat(fields[colSettlementAmount], 64)
		if err != nil {
			return fmt.Errorf("parse amount %s error: %v", fields[colSettlementAmount], err)
		}
		irO.Money = amount
		irO.Type = ir.TypeUnknown
		irO.Metadata["account_number"] = fields[colAccountNumber]
	case "证券转银行":
		irO.OrderType = ir.OrderTypeChinaSecuritiesBrokerTransferToBank
		amount, err := strconv.ParseFloat(fields[colSettlementAmount], 64)
		if err != nil {
			return fmt.Errorf("parse amount %s error: %v", fields[colSettlementAmount], err)
		}
		irO.Money = amount
		irO.Type = ir.TypeUnknown
		irO.Metadata["account_number"] = fields[colAccountNumber]
	case "利息归本":
		irO.OrderType = ir.OrderTypeChinaSecuritiesInterestCapitalization
		amount, err := strconv.ParseFloat(fields[colSettlementAmount], 64)
		if err != nil {
			return fmt.Errorf("parse amount %s error: %v", fields[colSettlementAmount], err)
		}
		irO.Money = amount
		irO.Type = ir.TypeUnknown
		irO.Metadata["account_number"] = fields[colAccountNumber]
	case "红利入账":
		// The transaction happens at end of the day
		fields[colTradeTime] = "23:59:59"

		irO.OrderType = ir.OrderTypeChinaSecuritiesDividend
		amount, err := strconv.ParseFloat(fields[colSettlementAmount], 64)
		if err != nil {
			return fmt.Errorf("parse amount %s error: %v", fields[colSettlementAmount], err)
		}
		irO.Money = amount
		irO.Type = ir.TypeRecv
		code := fmt.Sprintf("%06s", fields[colSecurityCode])
		irO.Item = "红利入账-" + code + "-" + fields[colSecurityName]

	case "证券买入", "证券卖出":
		irO.OrderType = ir.OrderTypeSecuritiesTrade
		volume, err := strconv.ParseInt(fields[colTradeVolume], 10, 64)
		if err != nil {
			return fmt.Errorf("parse volume %s error: %v", fields[colTradeVolume], err)
		}
		price, err := strconv.ParseFloat(fields[colTradePrice], 64)
		if err != nil {
			return fmt.Errorf("parse price %s error: %v", fields[colTradePrice], err)
		}
		commission, err := strconv.ParseFloat(fields[colTotalFee], 64)
		if err != nil {
			return fmt.Errorf("parse commission %s error: %v", fields[colTotalFee], err)
		}

		code := fmt.Sprintf("%06s", fields[colSecurityCode])
		txTypeOriginal := getTransactionType(fields[colShareholderCode], code)

		irO.TxTypeOriginal = txTypeOriginal
		irO.Item = code + "-" + fields[colSecurityName]
		irO.Amount = float64(volume)
		irO.Price = price
		irO.Commission = commission
		irO.Money, _ = strconv.ParseFloat(fields[colTradeAmount], 64)

		if fields[colBusinessType] == "证券买入" {
			irO.Type = ir.TypeSend
		} else {
			irO.Type = ir.TypeRecv
		}

	case "融券回购", "融券购回":
		irO.OrderType = ir.OrderTypeSecuritiesTrade
		volume, err := strconv.ParseInt(fields[colTradeVolume], 10, 64)
		if err != nil {
			return fmt.Errorf("parse volume %s error: %v", fields[colTradeVolume], err)
		}
		interest_rate, err := strconv.ParseFloat(fields[colTradePrice], 64)
		if err != nil {
			return fmt.Errorf("parse interest_rate %s error: %v", fields[colTradePrice], err)
		}
		commission, err := strconv.ParseFloat(fields[colTotalFee], 64)
		if err != nil {
			return fmt.Errorf("parse commission %s error: %v", fields[colTotalFee], err)
		}
		price := 100.0

		code := fmt.Sprintf("%06s", fields[colSecurityCode])
		txTypeOriginal := getTransactionType(fields[colShareholderCode], code)

		irO.TxTypeOriginal = txTypeOriginal
		irO.Item = code + "-" + fields[colSecurityName]
		irO.Amount = float64(volume)
		irO.Price = price
		irO.Commission = commission
		irO.Money, _ = strconv.ParseFloat(fields[colTradeAmount], 64)

		if fields[colBusinessType] == "融券回购" {
			irO.Type = ir.TypeSend
		} else {
			irO.Type = ir.TypeRecv
		}

		irO.Metadata["interest_rate"] = fmt.Sprintf("%.3f%%", interest_rate)

	case "ETF份额合并":
		irO.OrderType = ir.OrderTypeChinaSecuritiesEtfMerge

		// The transaction happens at end of the day
		fields[colTradeTime] = "23:59:59"

		// Example: 600 shares become 60 shares (10-for-1 reverse split)
		// 成交数量 (colTradeVolume) = 540 (reduction)
		// 剩余数量 (colRemainingVolume) = 60 (final amount)
		// Initial amount = 540 + 60 = 600
		tradeVolume, err := strconv.ParseFloat(fields[colTradeVolume], 64)
		if err != nil {
			return fmt.Errorf("parse trade volume %s error: %v", fields[colTradeVolume], err)
		}
		remainingVolume, err := strconv.ParseFloat(fields[colRemainingVolume], 64)
		if err != nil {
			return fmt.Errorf("parse remaining volume %s error: %v", fields[colRemainingVolume], err)
		}
		removedAmount := tradeVolume + remainingVolume // The original amount that is removed
		addedAmount := remainingVolume                 // The new amount after merge/split

		code := fmt.Sprintf("%06s", fields[colSecurityCode])
		txTypeOriginal := getTransactionType(fields[colShareholderCode], code)

		irO.TxTypeOriginal = txTypeOriginal
		irO.Item = code + "-" + fields[colSecurityName]
		irO.Amount = removedAmount                                    // Store the removed amount here
		irO.Type = ir.TypeUnknown                                     // No cash flow direction
		irO.Metadata["new_amount"] = fmt.Sprintf("%.2f", addedAmount) // Store added amount in metadata
		irO.Metadata["shareholder_code"] = fields[colShareholderCode]
		irO.Metadata["account_number"] = fields[colAccountNumber]
		// No trade_id for this type

	case "指定交易":
		// This transaction type doesn't generate entries.
		return nil
	default:
		return fmt.Errorf("unknown transaction type: %s", fields[colBusinessType])
	}

	// Handle trade time
	irO.Metadata["trade_time"] = fields[colTradeTime]
	payTime, err := time.Parse(LocalTimeFmt, fields[colSettlementDate][0:4]+"-"+fields[colSettlementDate][4:6]+"-"+fields[colSettlementDate][6:8]+" "+fields[colTradeTime]+" +0800")
	if err != nil {
		return fmt.Errorf("parse create time %s error: %v", fields[0], err)
	}
	irO.PayTime = payTime

	h.ir.Orders = append(h.ir.Orders, irO)
	return nil
}

func (h *Hxsec) Translate(filename string) (*ir.IR, error) {
	log.SetPrefix("[Provider-Hxsec] ")

	tradeReader, err := reader.GetGBKReader(filename)
	if err != nil {
		return nil, fmt.Errorf("can't get bill reader, err: %v", err)
	}

	scanner := bufio.NewScanner(tradeReader)
	for scanner.Scan() {
		h.LineNum++
		line := scanner.Text()
		if h.LineNum <= 1 {
			// The first line is file header.
			continue
		}

		fields := strings.Split(line, "\t")
		for i := range fields {
			fields[i] = strings.TrimPrefix(strings.TrimSuffix(fields[i], "\""), "=\"")
		}

		if err := h.fieldsToIR(fields); err != nil {
			return nil, err
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("Error reading file: %v", err)
	}

	log.Printf("Finished to parse the file %s", filename)
	return h.ir, nil
}

func getTransactionType(shareholderCode, code string) string {
	if strings.HasPrefix(shareholderCode, "A") {
		return "SH" + code
	}
	return "SZ" + code
}
