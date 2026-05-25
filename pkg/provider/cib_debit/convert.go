package cib_debit

import (
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
)

const fxMatchWindow = 3 * time.Minute

func (c *CibDebit) convertToIR() (*ir.IR, error) {
	i := ir.New()
	for _, order := range c.Orders {
		payTime, err := parseTradeTime(order.TradeTime, order.AccountingDay)
		if err != nil {
			return nil, err
		}

		money, txType, err := parseMoneyAndType(order.Expense, order.Income)
		if err != nil {
			return nil, err
		}
		if txType == OrderTypeUnknown || money == 0 {
			continue
		}

		irOrder := ir.Order{
			OrderType:      ir.OrderTypeNormal,
			Peer:           normalizePeer(order.Peer, order.PeerBank, order.PeerAccount),
			Item:           normalizeItem(order.Summary, order.Purpose, order.Remark),
			PayTime:        payTime,
			Type:           convertType(txType),
			TypeOriginal:   string(txType),
			TxTypeOriginal: order.Summary,
			Money:          money,
			Currency:       order.Currency,
			Metadata:       c.getMetadata(order),
		}
		i.Orders = append(i.Orders, irOrder)
	}
	i.Orders = c.mergeCurrencyExchangeOrders(i.Orders)
	return i, nil
}

func convertType(t OrderType) ir.Type {
	switch t {
	case OrderTypeRecv:
		return ir.TypeRecv
	case OrderTypeSend:
		return ir.TypeSend
	default:
		return ir.TypeUnknown
	}
}

func (c *CibDebit) getMetadata(order Order) map[string]string {
	metadata := map[string]string{
		"source":        providerSource,
		"accountName":   order.AccountName,
		"accountNum":    order.AccountNum,
		"subAccount":    order.SubAccount,
		"currency":      order.Currency,
		"accountingDay": order.AccountingDay,
		"summary":       order.Summary,
	}
	optional := map[string]string{
		"expense":     order.Expense,
		"income":      order.Income,
		"balance":     order.Balance,
		"peer":        order.Peer,
		"peerBank":    order.PeerBank,
		"peerAccount": order.PeerAccount,
		"purpose":     order.Purpose,
		"channel":     order.Channel,
		"remark":      order.Remark,
	}
	for key, value := range optional {
		if value != "" {
			metadata[key] = value
		}
	}
	return metadata
}

func (c *CibDebit) mergeCurrencyExchangeOrders(orders []ir.Order) []ir.Order {
	used := make([]bool, len(orders))
	merged := make([]ir.Order, 0, len(orders))

	for idx := range orders {
		if used[idx] {
			continue
		}
		source := orders[idx]
		if !isCurrencyExchangeCandidate(source) || source.Type != ir.TypeSend {
			continue
		}

		matchIdx := findBestCurrencyExchangeMatch(source, orders, used, idx)
		if matchIdx < 0 {
			continue
		}

		target := orders[matchIdx]
		used[idx] = true
		used[matchIdx] = true
		exchange := buildCurrencyExchangeOrder(source, target)
		log.Printf(
			"自动合并%s交易: %s %.2f %s -> %.2f %s",
			source.TxTypeOriginal,
			source.PayTime.Format(dateTimeLayout),
			source.Money,
			source.Currency,
			target.Money,
			target.Currency,
		)
		merged = append(merged, exchange)
	}

	for idx := range orders {
		if !used[idx] {
			merged = append(merged, orders[idx])
		}
	}

	sort.SliceStable(merged, func(i, j int) bool {
		return merged[i].PayTime.Before(merged[j].PayTime)
	})
	return merged
}

func findBestCurrencyExchangeMatch(source ir.Order, orders []ir.Order, used []bool, sourceIdx int) int {
	bestIdx := -1
	bestDiff := fxMatchWindow + time.Nanosecond

	for idx, target := range orders {
		if idx == sourceIdx || used[idx] {
			continue
		}
		if !isCurrencyExchangeCandidate(target) || target.Type != ir.TypeRecv {
			continue
		}
		if source.Currency == target.Currency {
			continue
		}
		if source.Metadata["accountNum"] == "" || source.Metadata["accountNum"] != target.Metadata["accountNum"] {
			continue
		}
		if source.TxTypeOriginal != target.TxTypeOriginal {
			continue
		}

		diff := source.PayTime.Sub(target.PayTime)
		if diff < 0 {
			diff = -diff
		}
		if diff <= fxMatchWindow && diff < bestDiff {
			bestDiff = diff
			bestIdx = idx
		}
	}

	return bestIdx
}

func isCurrencyExchangeCandidate(order ir.Order) bool {
	return order.TxTypeOriginal == "购汇" || order.TxTypeOriginal == "结汇"
}

func buildCurrencyExchangeOrder(source, target ir.Order) ir.Order {
	metadata := map[string]string{
		"source":           providerSource,
		"accountNum":       source.Metadata["accountNum"],
		"sourceSubAccount": source.Metadata["subAccount"],
		"targetSubAccount": target.Metadata["subAccount"],
		"sourceCurrency":   source.Currency,
		"targetCurrency":   target.Currency,
		"sourceAmount":     formatAmount(source.Money),
		"targetAmount":     formatAmount(target.Money),
		"matchWindow":      fxMatchWindow.String(),
	}
	if source.Metadata["accountName"] != "" {
		metadata["accountName"] = source.Metadata["accountName"]
	}

	return ir.Order{
		OrderType:      ir.OrderTypeCurrencyExchange,
		Peer:           providerPeer,
		Item:           source.TxTypeOriginal + " " + source.Currency + "/" + target.Currency,
		PayTime:        source.PayTime,
		Type:           ir.TypeSend,
		TypeOriginal:   source.TypeOriginal,
		TxTypeOriginal: source.TxTypeOriginal,
		Money:          source.Money,
		Amount:         target.Money,
		Currency:       source.Currency,
		Units: map[ir.Unit]string{
			ir.TargetUnit: target.Currency,
		},
		Metadata: metadata,
	}
}

func formatAmount(amount float64) string {
	return strconv.FormatFloat(amount, 'f', 2, 64)
}
