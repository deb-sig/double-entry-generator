package alipay

import (
	"fmt"

	"github.com/gaocegege/double-entry-generator/pkg/ir"
)

// convertToIR convert alipay bills to IR.
func (a *Alipay) convertToIR() *ir.IR {
	i := ir.New()
	for _, o := range a.Orders {

		irO := ir.Order{
			Peer:           o.Peer,
			Item:           o.ItemName,
			Method:         o.Method,
			Note:           fmt.Sprintf("%s-%s-%s", o.TxTypeOriginal, o.Status, o.Category),
			PayTime:        o.PayTime,
			Money:          o.Money,
			OrderID:        &o.DealNo,
			TxType:         conevertType(o.TxType),
			TxTypeOriginal: o.TxTypeOriginal,
		}
		if o.MerchantId != "" {
			irO.MerchantOrderID = &o.MerchantId
		}
		i.Orders = append(i.Orders, irO)
	}
	return i
}

func conevertType(t TxTypeType) ir.TxType {
	switch t {
	case TxTypeSend:
		return ir.TxTypeSend
	case TxTypeRecv:
		return ir.TxTypeRecv
	default:
		return ir.TxTypeUnknown
	}
}
