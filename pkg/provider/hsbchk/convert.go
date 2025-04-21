package hsbchk

import "github.com/deb-sig/double-entry-generator/pkg/ir"

// convertToIR 将解析后的订单转换为中间表示
func (h *HsbcHK) convertToIR() *ir.IR {
	i := ir.New()
	for _, o := range h.Orders {
		irO := ir.Order{
			OrderType:    ir.OrderTypeNormal,
			PayTime:      o.PayTime,
			Type:         convertOrderType(o.Type),
			TypeOriginal: string(o.Type),
			Peer:         o.Merchant, // 使用商户名称作为交易对方
			Item:         o.Description,
			Money:        o.Money,
		}

		// 根据卡片模式设置额外信息
		if h.Mode == DebitMode {
			// 借记卡特有信息
		} else {
			// 信用卡特有信息
			irO.TxTypeOriginal = o.StatusOriginal
		}

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
