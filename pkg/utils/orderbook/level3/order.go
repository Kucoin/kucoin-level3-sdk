package level3

import (
	"fmt"

	"github.com/Kucoin/kucoin-level3-sdk/pkg/utils/orderbook/base"
	"github.com/shopspring/decimal"
)

type Order struct {
	OrderId string
	Side    string
	Price   decimal.Decimal
	Size    decimal.Decimal
	Time    uint64
	Info    interface{}
}

func NewOrder(orderId string, side string, price string, size string, time uint64, info interface{}) (order *Order, err error) {
	if err := base.CheckSide(side); err != nil {
		return nil, err
	}

	priceValue, err := decimal.NewFromString(price)
	if err != nil {
		return nil, fmt.Errorf("NewOrder failed, price: `%s`, error: %v", price, err)
	}

	sizeValue, err := decimal.NewFromString(size)
	if err != nil {
		return nil, fmt.Errorf("NewOrder failed, size: `%s`, error: %v", size, err)
	}

	order = &Order{
		OrderId: orderId,
		Side:    side,
		Price:   priceValue,
		Size:    sizeValue,
		Time:    time,
		Info:    info,
	}
	return
}
