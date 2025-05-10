package hsbchk

import (
	"fmt"

	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
)

// convertToIR 将解析后的订单转换为中间表示
func (h *HsbcHK) convertToIR() *ir.IR {
	i := ir.New()
	for _, o := range h.Orders {
		irO := ir.Order{
			OrderType: ir.OrderTypeNormal,
			PayTime:   o.PayTime,
			Type:      convertOrderType(o.Type),
			Money:     o.Money,
			Currency:  o.Currency,
		}

		// 根据卡片模式设置额外信息
		if h.Mode == DebitMode {
			// 借记卡特有信息
			irO.Peer = o.Description
		} else {
			// 信用卡特有信息
			irO.TxTypeOriginal = o.StatusOriginal
			irO.Peer = o.Merchant
			irO.Item = o.Description
		}

		irO.Metadata = h.getMetadata(o)
		i.Orders = append(i.Orders, irO)
	}
	return i
}

// convertOrderType 将HSBC HK的订单类型转换为IR的类型
func convertOrderType(t OrderType) ir.Type {
	switch t {
	case OrderTypeSend:
		return ir.TypeSend
	case OrderTypeRecv:
		return ir.TypeRecv
	default:
		return ir.TypeUnknown
	}
}

// getMetadata get the metadata (e.g. status, method, category and so on.)
//
//	from order.
func (h *HsbcHK) getMetadata(o Order) map[string]string {
	// FIXME(TripleZ): hard-coded, bad pattern
	source := "HSBC HK"

	if h.Mode == DebitMode {
		// 借记卡特有信息
		var balance, balanceCurrency string
		if o.Balance != 0 {
			balance = fmt.Sprintf("%.2f", o.Balance)
		}
		if o.BalanceCurrency != "" {
			balanceCurrency = o.BalanceCurrency
		}

		metadata := map[string]string{
			"source":  source,
			"balance": fmt.Sprintf("%s %s", balance, balanceCurrency),
		}

		return metadata
	}

	// 信用卡特有信息
	var postDate, country, creditDebit string
	if !o.PostDate.IsZero() {
		postDate = o.PostDate.Format(TimeFormat)
	}
	if o.Country != "" {
		country = o.Country
	}
	if o.CreditDebit != "" {
		creditDebit = o.CreditDebit
	}

	metadata := map[string]string{
		"source":          source,
		"post_date":       postDate,
		"country":         country,
		"credit_or_debit": creditDebit,
	}

	return metadata
}
