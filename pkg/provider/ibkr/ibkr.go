package ibkr

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/deb-sig/double-entry-generator/v2/pkg/io/reader"
	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
)

type Ibkr struct{}

func New() *Ibkr {
	return &Ibkr{}
}

func (i *Ibkr) Translate(filename string) (*ir.IR, error) {
	log.SetPrefix("[Provider-IBKR] ")

	billReader, err := reader.GetReader(filename)
	if err != nil {
		return nil, fmt.Errorf("can't get bill reader, err: %v", err)
	}

	decoder := xml.NewDecoder(billReader)
	result := ir.New()
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		start, ok := token.(xml.StartElement)
		if !ok {
			continue
		}

		attrs := attrsToMap(start.Attr)
		switch start.Name.Local {
		case "Trade":
			order, ok, err := tradeToOrder(attrs)
			if err != nil {
				return nil, err
			}
			if ok {
				result.Orders = append(result.Orders, order)
			}
		case "CashTransaction":
			order, ok, err := cashTransactionToOrder(attrs)
			if err != nil {
				return nil, err
			}
			if ok {
				result.Orders = append(result.Orders, order)
			}
		}
	}

	log.Printf("Finished to parse the file %s", filename)
	return result, nil
}

func attrsToMap(attrs []xml.Attr) map[string]string {
	result := make(map[string]string, len(attrs))
	for _, attr := range attrs {
		result[attr.Name.Local] = strings.TrimSpace(attr.Value)
	}
	return result
}

func tradeToOrder(attrs map[string]string) (ir.Order, bool, error) {
	switch attrs["assetCategory"] {
	case "STK":
		return stockTradeToOrder(attrs)
	case "CASH":
		return cashTradeToOrder(attrs)
	default:
		return ir.Order{}, false, nil
	}
}

func stockTradeToOrder(attrs map[string]string) (ir.Order, bool, error) {
	payTime, err := parseDate(attrs["tradeDate"], attrs["dateTime"])
	if err != nil {
		return ir.Order{}, false, err
	}
	quantity, err := parseRequiredFloat(attrs, "quantity")
	if err != nil {
		return ir.Order{}, false, err
	}
	price, err := parseRequiredFloat(attrs, "tradePrice")
	if err != nil {
		return ir.Order{}, false, err
	}
	money, err := parseRequiredFloat(attrs, "tradeMoney")
	if err != nil {
		return ir.Order{}, false, err
	}
	commission, err := parseOptionalFloat(attrs["ibCommission"])
	if err != nil {
		return ir.Order{}, false, fmt.Errorf("parse ibCommission %q error: %v", attrs["ibCommission"], err)
	}

	orderType, err := convertBuySell(attrs["buySell"])
	if err != nil {
		return ir.Order{}, false, err
	}

	symbol := commodityName(attrs["symbol"])
	description := attrs["description"]
	item := symbol
	if description != "" {
		item = symbol + "-" + description
	}

	return ir.Order{
		OrderType:      ir.OrderTypeSecuritiesTrade,
		Peer:           "IBKR",
		PayTime:        payTime,
		TxTypeOriginal: symbol,
		Type:           orderType,
		TypeOriginal:   attrs["buySell"],
		Item:           item,
		Money:          math.Abs(money),
		Amount:         math.Abs(quantity),
		Price:          price,
		Commission:     math.Abs(commission),
		Currency:       attrs["currency"],
		Metadata:       tradeMetadata(attrs),
	}, true, nil
}

func cashTradeToOrder(attrs map[string]string) (ir.Order, bool, error) {
	payTime, err := parseDate(attrs["tradeDate"], attrs["dateTime"])
	if err != nil {
		return ir.Order{}, false, err
	}
	quantity, err := parseRequiredFloat(attrs, "quantity")
	if err != nil {
		return ir.Order{}, false, err
	}
	tradeMoney, err := parseRequiredFloat(attrs, "tradeMoney")
	if err != nil {
		return ir.Order{}, false, err
	}
	orderType, err := convertBuySell(attrs["buySell"])
	if err != nil {
		return ir.Order{}, false, err
	}

	baseCurrency, quoteCurrency := splitForexSymbol(attrs["symbol"], attrs["currency"])
	sourceCurrency := quoteCurrency
	sourceAmount := math.Abs(tradeMoney)
	targetCurrency := baseCurrency
	targetAmount := math.Abs(quantity)
	if orderType == ir.TypeRecv {
		sourceCurrency = baseCurrency
		sourceAmount = math.Abs(quantity)
		targetCurrency = quoteCurrency
		targetAmount = math.Abs(tradeMoney)
	}

	return ir.Order{
		OrderType: ir.OrderTypeCurrencyExchange,
		Peer:      "IBKR",
		PayTime:   payTime,
		Item:      "Forex " + attrs["symbol"],
		Money:     sourceAmount,
		Currency:  sourceCurrency,
		Amount:    targetAmount,
		Units: map[ir.Unit]string{
			ir.TargetUnit: targetCurrency,
		},
		Metadata: tradeMetadata(attrs),
	}, true, nil
}

func cashTransactionToOrder(attrs map[string]string) (ir.Order, bool, error) {
	amount, err := parseRequiredFloat(attrs, "amount")
	if err != nil {
		return ir.Order{}, false, err
	}
	if amount == 0 {
		return ir.Order{}, false, nil
	}
	payTime, err := parseDate(attrs["settleDate"], attrs["dateTime"])
	if err != nil {
		return ir.Order{}, false, err
	}

	orderType := ir.TypeRecv
	if amount < 0 {
		orderType = ir.TypeSend
	}
	item := attrs["type"]
	if attrs["description"] != "" {
		item = item + "-" + attrs["description"]
	}

	return ir.Order{
		OrderType:      ir.OrderTypeNormal,
		Peer:           "IBKR",
		PayTime:        payTime,
		Type:           orderType,
		TypeOriginal:   attrs["type"],
		TxTypeOriginal: attrs["type"],
		Item:           item,
		Money:          math.Abs(amount),
		Currency:       attrs["currency"],
		Metadata:       cashTransactionMetadata(attrs),
	}, true, nil
}

func convertBuySell(value string) (ir.Type, error) {
	switch value {
	case "BUY":
		return ir.TypeSend, nil
	case "SELL":
		return ir.TypeRecv, nil
	default:
		return ir.TypeUnknown, fmt.Errorf("unsupported IBKR buySell %q", value)
	}
}

func parseRequiredFloat(attrs map[string]string, key string) (float64, error) {
	value, ok := attrs[key]
	if !ok || value == "" {
		return 0, fmt.Errorf("missing %s", key)
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("parse %s %q error: %v", key, value, err)
	}
	return parsed, nil
}

func parseOptionalFloat(value string) (float64, error) {
	if value == "" {
		return 0, nil
	}
	return strconv.ParseFloat(value, 64)
}

func parseDate(dateValue, dateTimeValue string) (time.Time, error) {
	if dateValue == "" && len(dateTimeValue) >= len("2006-01-02") {
		dateValue = dateTimeValue[:len("2006-01-02")]
	}
	if dateValue == "" {
		return time.Time{}, fmt.Errorf("missing date")
	}
	parsed, err := time.Parse("2006-01-02", dateValue)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse date %q error: %v", dateValue, err)
	}
	return parsed, nil
}

func splitForexSymbol(symbol, fallbackQuote string) (string, string) {
	parts := strings.Split(symbol, ".")
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1]
	}
	return commodityName(symbol), fallbackQuote
}

func commodityName(symbol string) string {
	replacer := strings.NewReplacer(" ", "_", "/", "_", ":", "_")
	return replacer.Replace(strings.TrimSpace(symbol))
}

func tradeMetadata(attrs map[string]string) map[string]string {
	return pickMetadata(attrs, map[string]string{
		"accountId":            "account_id",
		"assetCategory":        "asset_category",
		"subCategory":          "sub_category",
		"tradeID":              "trade_id",
		"transactionID":        "transaction_id",
		"ibOrderID":            "ib_order_id",
		"ibExecID":             "ib_exec_id",
		"exchange":             "exchange",
		"isin":                 "isin",
		"securityID":           "security_id",
		"reportDate":           "report_date",
		"settleDateTarget":     "settle_date",
		"ibCommissionCurrency": "commission_currency",
	})
}

func cashTransactionMetadata(attrs map[string]string) map[string]string {
	return pickMetadata(attrs, map[string]string{
		"accountId":     "account_id",
		"assetCategory": "asset_category",
		"subCategory":   "sub_category",
		"symbol":        "symbol",
		"transactionID": "transaction_id",
		"actionID":      "action_id",
		"isin":          "isin",
		"securityID":    "security_id",
		"reportDate":    "report_date",
		"settleDate":    "settle_date",
		"dateTime":      "date_time",
	})
}

func pickMetadata(attrs map[string]string, keys map[string]string) map[string]string {
	result := map[string]string{}
	for source, target := range keys {
		if attrs[source] != "" {
			result[target] = attrs[source]
		}
	}
	return result
}
