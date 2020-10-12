package level3

import (
	"encoding/json"
	"fmt"

	"github.com/Kucoin/kucoin-level3-sdk/pkg/utils/orderbook/base"
	"github.com/Kucoin/kucoin-level3-sdk/pkg/utils/orderbook/skiplist"
	"github.com/shopspring/decimal"
)

type OrderBook struct {
	Sequence  uint64
	Asks      *skiplist.SkipList //Sort price from low to high
	Bids      *skiplist.SkipList //Sort price from high to low
	orderPool map[string]*Order
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		Asks:      newAskOrders(),
		Bids:      newBidOrders(),
		orderPool: make(map[string]*Order),
	}
}

func isEqual(l, r interface{}) bool {
	switch val := l.(type) {
	case decimal.Decimal:
		cVal := r.(decimal.Decimal)
		if !val.Equals(cVal) {
			return false
		}

	case *Order:
		cVal := r.(*Order)
		if cVal.OrderId != val.OrderId {
			return false
		}

	default:
		if val != r {
			return false
		}
	}
	return true
}

func newAskOrders() *skiplist.SkipList {
	return skiplist.NewCustomMap(func(l, r interface{}) bool {
		if l.(*Order).Price.Equal(r.(*Order).Price) {
			return l.(*Order).Time < r.(*Order).Time
		}

		return l.(*Order).Price.LessThan(r.(*Order).Price)
	}, isEqual)
}

func newBidOrders() *skiplist.SkipList {
	return skiplist.NewCustomMap(func(l, r interface{}) bool {
		if l.(*Order).Price.Equal(r.(*Order).Price) {
			return l.(*Order).Time < r.(*Order).Time
		}

		return l.(*Order).Price.GreaterThan(r.(*Order).Price)
	}, isEqual)
}

func (ob *OrderBook) getOrderBookBySide(side string) (*skiplist.SkipList, error) {
	if err := base.CheckSide(side); err != nil {
		return nil, err
	}

	if side == base.AskSide {
		return ob.Asks, nil
	}

	return ob.Bids, nil
}

func (ob *OrderBook) AddOrder(order *Order) error {
	orderBook, err := ob.getOrderBookBySide(order.Side)
	if err != nil {
		return err
	}

	orderBook.Set(order, order)
	ob.orderPool[order.OrderId] = order
	return nil
}

func (ob *OrderBook) RemoveByOrderId(orderId string) error {
	order, ok := ob.orderPool[orderId]
	if !ok {
		return nil
	}

	if err := ob.removeOrder(order); err != nil {
		return err
	}
	return nil
}

func (ob *OrderBook) removeOrder(order *Order) error {
	orderBook, err := ob.getOrderBookBySide(order.Side)
	if err != nil {
		return err
	}
	if _, ok := orderBook.Delete(order); ok {
		delete(ob.orderPool, order.OrderId)
	}

	return nil
}

func (ob *OrderBook) GetOrder(orderId string) *Order {
	order, ok := ob.orderPool[orderId]
	if !ok {
		return nil
	}

	return order
}

func (ob *OrderBook) MatchOrder(orderId string, size decimal.Decimal) error {
	order, ok := ob.orderPool[orderId]
	if !ok {
		return nil
	}

	newSize := order.Size.Sub(size)
	if newSize.LessThan(decimal.Zero) {
		return fmt.Errorf("oldSize: %s, size: %s, sub result less than zero", order.Size.String(), size)
	}

	order.Size = newSize
	if order.Size.Equal(decimal.Zero) {
		if err := ob.removeOrder(order); err != nil {
			return err
		}
		return nil
	}

	return nil
}

func (ob *OrderBook) ChangeOrder(orderId string, size decimal.Decimal) error {
	order, ok := ob.orderPool[orderId]
	if !ok {
		return nil
	}

	order.Size = size
	if order.Size.Equal(decimal.Zero) {
		if err := ob.removeOrder(order); err != nil {
			return err
		}
		return nil
	}
	return nil
}

func (ob *OrderBook) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"sequence":   ob.Sequence,
		base.AskSide: ob.GetL3PartOrderBookBySide(base.AskSide, 0),
		base.BidSide: ob.GetL3PartOrderBookBySide(base.BidSide, 0),
	})
}

// Level3 OrderBook
func (ob *OrderBook) GetL3PartOrderBookBySide(side string, number int) [][3]string {
	if err := base.CheckSide(side); err != nil {
		return nil
	}

	var it skiplist.Iterator
	if side == base.AskSide {
		it = ob.Asks.Iterator()
		if number == 0 {
			number = ob.Asks.Len()
		} else {
			number = base.Min(number, ob.Asks.Len())
		}
	} else {
		it = ob.Bids.Iterator()
		if number == 0 {
			number = ob.Bids.Len()
		} else {
			number = base.Min(number, ob.Bids.Len())
		}
	}
	arr := make([][3]string, 0, number)

	for it.Next() {
		if len(arr) >= number {
			break
		}

		order := it.Value().(*Order)
		arr = append(arr, [3]string{order.OrderId, order.Price.String(), order.Size.String()})
	}

	return arr
}

// Level2 OrderBook
func (ob *OrderBook) GetPartOrderBookBySide(side string, number int) [][2]string {
	if err := base.CheckSide(side); err != nil {
		return nil
	}

	var it skiplist.Iterator
	if side == base.AskSide {
		it = ob.Asks.Iterator()
		if number == 0 {
			number = ob.Asks.Len()
		} else {
			number = base.Min(number, ob.Asks.Len())
		}
	} else {
		it = ob.Bids.Iterator()
		if number == 0 {
			number = ob.Bids.Len()
		} else {
			number = base.Min(number, ob.Bids.Len())
		}
	}
	arr := make([][2]string, 0, number)

	lastPrice := decimal.Zero
	lastPriceSize := decimal.Zero

	for it.Next() {
		if len(arr) >= number {
			break
		}

		order := it.Value().(*Order)
		if lastPrice.Equal(decimal.Zero) {
			lastPrice = order.Price
			lastPriceSize = order.Size
			continue
		}
		if lastPrice.Equal(order.Price) {
			lastPriceSize = lastPriceSize.Add(order.Size)
			continue
		}
		arr = append(arr, [2]string{lastPrice.String(), lastPriceSize.String()})
		lastPrice = order.Price
		lastPriceSize = order.Size
	}

	return arr
}

func (ob *OrderBook) GetOrderBookTickerOrder() (askOrder, bidOrder *Order) {
	askIT := ob.Asks.Iterator()
	askIT.Next()
	ask := askIT.Value()
	switch ask.(type) {
	case *Order:
		askOrder = ask.(*Order)
	}
	bidIT := ob.Bids.Iterator()
	bidIT.Next()
	bid := bidIT.Value()
	switch bid.(type) {
	case *Order:
		bidOrder = bid.(*Order)
	}
	return
}
