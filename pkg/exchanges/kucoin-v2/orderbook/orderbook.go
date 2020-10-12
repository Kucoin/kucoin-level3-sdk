package orderbook

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/Kucoin/kucoin-level3-sdk/pkg/consts"
	"github.com/Kucoin/kucoin-level3-sdk/pkg/exchanges"
	"github.com/Kucoin/kucoin-level3-sdk/pkg/exchanges/kucoin-v2/sdk"
	"github.com/Kucoin/kucoin-level3-sdk/pkg/exchanges/kucoin-v2/stream"
	"github.com/Kucoin/kucoin-level3-sdk/pkg/services/log"
	"github.com/Kucoin/kucoin-level3-sdk/pkg/utils/orderbook/base"
	"github.com/Kucoin/kucoin-level3-sdk/pkg/utils/orderbook/level3"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type Builder struct {
	apiService *sdk.Kucoin
	symbol     string
	lock       *sync.RWMutex
	Messages   chan *sdk.WebSocketDownstreamMessage

	OrderBookTime uint64
	Sequence      uint64 //Sequence || UpdateID
	fullOrderBook *level3.OrderBook
}

func NewBuilder(apiService *sdk.Kucoin, symbol string) *Builder {
	return &Builder{
		apiService: apiService,
		symbol:     symbol,
		lock:       &sync.RWMutex{},
		Messages:   make(chan *sdk.WebSocketDownstreamMessage, consts.MaxMsgChanLen),
	}
}

func (b *Builder) resetOrderBook() {
	b.lock.Lock()
	b.fullOrderBook = level3.NewOrderBook()
	b.lock.Unlock()
}

func (b *Builder) ReloadOrderBook() {
	defer func() {
		if r := recover(); r != nil {
			log.Panic("ReloadOrderBook Panic", zap.Any("recover", r))
		}
	}()

	log.Info("start running ReloadOrderBook, symbol: " + b.symbol)
	b.resetOrderBook()

	b.playback()

	for msg := range b.Messages {
		l3Data, err := stream.NewStreamDataModel(msg)
		if err != nil {
			log.Panic("NewStreamDataModel panic", zap.Error(err))
		}
		b.updateFromStream(l3Data)
	}
}

func (b *Builder) playback() {
	log.Info("prepare playback...")

	const tempMsgChanMaxLen = 10240
	tempMsgChan := make(chan *stream.DataModel, tempMsgChanMaxLen)
	firstSequence := uint64(0)
	var fullOrderBook *DepthResponse

	for msg := range b.Messages {
		l3Data, err := stream.NewStreamDataModel(msg)
		if err != nil {
			log.Panic("NewStreamDataModel panic", zap.Error(err))
		}
		tempMsgChan <- l3Data

		if firstSequence == 0 {
			firstSequence = l3Data.Sequence
			log.Info(fmt.Sprintf("firstSequence: %d", firstSequence))
		}

		if len(tempMsgChan) > 5 {
			if fullOrderBook == nil {
				log.Info("start GetAtomicFullOrderBook , symbol: " + b.symbol)
				fullOrderBook, err = b.GetAtomicFullOrderBook()
				if err != nil {
					log.Panic("GetAtomicFullOrderBook panic", zap.Error(err))
					continue
				}
				log.Info(fmt.Sprintf("finish GetAtomicFullOrderBook, Sequence: %d", fullOrderBook.Sequence))
			}

			if len(tempMsgChan) > tempMsgChanMaxLen-5 {
				log.Panic("playback failed, tempMsgChan is too long, retry...")
			}

			if fullOrderBook != nil && fullOrderBook.Sequence < firstSequence {
				log.Info(fmt.Sprintf("fullOrderBook Sequence is small, Sequence: %d", fullOrderBook.Sequence))
				fullOrderBook = nil
				continue
			}

			if fullOrderBook != nil && fullOrderBook.Sequence <= l3Data.Sequence { //string camp
				log.Info("sequence match, start playback, tempMsgChan: " + strconv.Itoa(len(tempMsgChan)))

				b.lock.Lock()
				b.AddDepthToOrderBook(fullOrderBook)
				b.lock.Unlock()

				n := len(tempMsgChan)
				for i := 0; i < n; i++ {
					b.updateFromStream(<-tempMsgChan)
				}

				log.Info("finish playback.")
				break
			}
		}
	}
}

func newOrderWithElem(side string, elem [4]interface{}, info interface{}) (*level3.Order, error) {
	timeInt, err := elem[3].(json.Number).Int64()
	if err != nil {
		return nil, err
	}
	return level3.NewOrder(elem[0].(string), side, elem[1].(string), elem[2].(string), uint64(timeInt), info)
}

func (b *Builder) AddDepthToOrderBook(depth *DepthResponse) {
	b.Sequence = depth.Sequence
	b.OrderBookTime = uint64(time.Now().UnixNano())
	b.formatDepthToOrderBook(depth, b.fullOrderBook)
}

func (b *Builder) formatDepthToOrderBook(depth *DepthResponse, fullOrderBook *level3.OrderBook) {
	fullOrderBook.Sequence = depth.Sequence

	for _, elem := range depth.Asks {
		order, err := newOrderWithElem(base.AskSide, elem, nil)
		if err != nil {
			log.Panic("NewOrder panic", zap.Error(err))
		}

		if err := fullOrderBook.AddOrder(order); err != nil {
			log.Panic("AddOrder panic", zap.Error(err))
		}
	}

	for _, elem := range depth.Bids {
		order, err := newOrderWithElem(base.BidSide, elem, nil)
		if err != nil {
			log.Panic("NewOrder panic", zap.Error(err))
		}

		if err := fullOrderBook.AddOrder(order); err != nil {
			log.Panic("AddOrder panic", zap.Error(err))
		}
	}

	return
}

func (b *Builder) updateFromStream(msg *stream.DataModel) {
	b.lock.Lock()
	defer b.lock.Unlock()

	skip, err := b.updateSequence(msg)
	if err != nil {
		log.Panic("updateSequence panic", zap.Error(err))
	}

	if !skip {
		b.updateOrderBook(msg)
	}
}

func (b *Builder) updateSequence(msg *stream.DataModel) (bool, error) {
	if b.Sequence+1 > msg.Sequence {
		return true, nil
	}

	if b.Sequence+1 != msg.Sequence {
		return false, errors.New(fmt.Sprintf(
			"currentSequence: %d, msgSequence: %d, the sequence is not continuous, current chanLen: %d",
			b.Sequence,
			msg.Sequence,
			len(b.Messages),
		))
	}

	b.Sequence = msg.Sequence
	b.fullOrderBook.Sequence = msg.Sequence
	return false, nil
}

func (b *Builder) updateOrderBook(msg *stream.DataModel) {
	//[3]string{"orderId", "price", "size"}
	//var item = [3]string{msg.OrderId, msg.Price, msg.Size}

	switch msg.Type {
	case stream.MessageReceivedType:

	case stream.MessageOpenType:
		data := &stream.DataOpenModel{}
		if err := json.Unmarshal(msg.Data(), data); err != nil {
			log.Panic("Unmarshal panic", zap.Error(err))
		}

		if data.Price == "" || data.Size == "0" || data.Price == "0" || data.Size == "" {
			return
		}

		side := ""
		switch data.Side {
		case stream.SellSide:
			side = base.AskSide
		case stream.BuySide:
			side = base.BidSide
		default:
			panic("error side: " + data.Side)
		}

		order, err := level3.NewOrder(data.OrderId, side, data.Price, data.Size, data.Time, nil)
		if err != nil {
			log.Panic("NewOrder panic: "+err.Error(), zap.String("Data", string(msg.Data())))
		}
		if err := b.fullOrderBook.AddOrder(order); err != nil {
			log.Panic("AddOrder panic: " + err.Error())
		}
		b.OrderBookTime = data.Time

	case stream.MessageDoneType:
		data := &stream.DataDoneModel{}
		if err := json.Unmarshal(msg.Data(), data); err != nil {
			log.Panic("Unmarshal panic", zap.Error(err))
		}

		if err := b.fullOrderBook.RemoveByOrderId(data.OrderId); err != nil {
			log.Panic("RemoveByOrderId panic: " + err.Error())
		}
		b.OrderBookTime = data.Time

	case stream.MessageMatchType:
		data := &stream.DataMatchModel{}
		if err := json.Unmarshal(msg.Data(), data); err != nil {
			log.Panic("Unmarshal panic: " + err.Error())
		}
		size, err := decimal.NewFromString(data.RemainSize)
		if err != nil {
			log.Panic("MatchOrder panic: " + err.Error())
		}
		if err := b.fullOrderBook.ChangeOrder(data.MakerOrderId, size); err != nil {
			log.Panic("MatchOrder panic: " + err.Error())
		}
		b.OrderBookTime = data.Time

	case stream.MessageUpdateType:
		data := &stream.DataUpdateModel{}
		if err := json.Unmarshal(msg.Data(), data); err != nil {
			log.Panic("Unmarshal panic: " + err.Error())
		}

		size, err := decimal.NewFromString(data.Size)
		if err != nil {
			log.Panic("UpdateOrder panic: " + err.Error())
		}
		if err := b.fullOrderBook.ChangeOrder(data.OrderId, size); err != nil {
			log.Panic("UpdateOrder panic: " + err.Error())
		}
		b.OrderBookTime = data.Time

	default:
		log.Panic("error msg type: " + msg.Type)
	}

	ask, bid := b.fullOrderBook.GetOrderBookTickerOrder()
	if ask != nil && bid != nil && bid.Price.Cmp(ask.Price) >= 0 {
		log.Panic("order book cross", zap.String("asks", ask.Price.String()), zap.String("bids", bid.Price.String()))
	}
}

//[3]string{"orderId", "price", "size"}
type FullOrderBook struct {
	Sequence uint64      `json:"sequence"`
	Asks     [][3]string `json:"asks"`
	Bids     [][3]string `json:"bids"`
}

func (b *Builder) DepthResponse2FullOrderBook(atomicFullOrderBook *DepthResponse) (*FullOrderBook, error) {
	orderBook := level3.NewOrderBook()
	b.formatDepthToOrderBook(atomicFullOrderBook, orderBook)
	data, err := json.Marshal(orderBook)
	if err != nil {
		return nil, err
	}

	ret := &FullOrderBook{}
	if err := json.Unmarshal(data, ret); err != nil {
		return nil, err
	}

	return ret, nil
}

func (b *Builder) Snapshot() (*FullOrderBook, error) {
	data, err := b.SnapshotBytes()
	if err != nil {
		return nil, err
	}

	ret := &FullOrderBook{}
	if err := json.Unmarshal(data, ret); err != nil {
		return nil, err
	}

	return ret, nil
}

func (b *Builder) SnapshotBytes() ([]byte, error) {
	b.lock.RLock()
	data, err := json.Marshal(b.fullOrderBook)
	b.lock.RUnlock()
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (b *Builder) GetPartOrderBook(number int) *exchanges.OrderBook {
	defer func() {
		if r := recover(); r != nil {
			log.Error("GetPartOrderBook panic", zap.Any("r", r))
		}
	}()

	b.lock.RLock()
	defer b.lock.RUnlock()

	data := &exchanges.OrderBook{
		Asks: b.fullOrderBook.GetPartOrderBookBySide(base.AskSide, number),
		Bids: b.fullOrderBook.GetPartOrderBookBySide(base.BidSide, number),
		Info: map[string]interface{}{
			"time": b.OrderBookTime,
		},
	}

	return data
}

func (b *Builder) GetL3PartOrderBook(number int) *exchanges.Level3OrderBook {
	defer func() {
		if r := recover(); r != nil {
			log.Error("GetPartOrderBook panic", zap.Any("r", r))
		}
	}()

	b.lock.RLock()
	defer b.lock.RUnlock()

	data := &exchanges.Level3OrderBook{
		Asks: b.fullOrderBook.GetL3PartOrderBookBySide(base.AskSide, number),
		Bids: b.fullOrderBook.GetL3PartOrderBookBySide(base.BidSide, number),
		Info: map[string]interface{}{
			"time": b.OrderBookTime,
		},
	}

	return data
}
