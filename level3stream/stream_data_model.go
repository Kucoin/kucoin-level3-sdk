package level3stream

import (
	"encoding/json"
)

type StreamDataModel struct {
	Sequence   string `json:"sequence"`
	Symbol     string `json:"symbol"`
	Type       string `json:"type"`
	Side       string `json:"side"`
	rawMessage json.RawMessage
}

func NewStreamDataModel(msgData json.RawMessage) (*StreamDataModel, error) {
	l3Data := &StreamDataModel{}

	if err := json.Unmarshal(msgData, l3Data); err != nil {
		return nil, err
	}
	l3Data.rawMessage = msgData

	return l3Data, nil
}

func (l3Data *StreamDataModel) GetRawMessage() json.RawMessage {
	return l3Data.rawMessage
}

const (
	BuySide  = "buy"
	SellSide = "sell"

	LimitOrderType  = "limit"
	MarketOrderType = "market"

	MessageDoneCanceled = "canceled"
	MessageDoneFilled   = "filled"

	MessageReceivedType = "received"
	MessageOpenType     = "open"
	MessageDoneType     = "done"
	MessageMatchType    = "match"
	MessageChangeType   = "change"
)

type StreamDataReceivedModel struct {
	OrderType string `json:"orderType"`
	Side      string `json:"side"`
	//Size      string `json:"size"`
	Price string `json:"price"`
	//Funds     string `json:"funds"`
	OrderId   string `json:"orderId"`
	Time      string `json:"time"`
	ClientOid string `json:"clientOid"`
}

type StreamDataOpenModel struct {
	Side    string `json:"side"`
	Size    string `json:"size"`
	OrderId string `json:"orderId"`
	Price   string `json:"price"`
	Time    string `json:"time"`
	//RemainSize string `json:"remainSize"`
}

type StreamDataDoneModel struct {
	Side    string `json:"side"`
	Size    string `json:"size"`
	Reason  string `json:"reason"`
	OrderId string `json:"orderId"`
	Price   string `json:"price"`
	Time    string `json:"time"`
}

type StreamDataMatchModel struct {
	Side         string `json:"side"`
	Size         string `json:"size"`
	Price        string `json:"price"`
	TakerOrderId string `json:"takerOrderId"`
	MakerOrderId string `json:"makerOrderId"`
	Time         string `json:"time"`
	TradeId      string `json:"tradeId"`
}

type StreamDataChangeModel struct {
	Side    string `json:"side"`
	NewSize string `json:"newSize"`
	OldSize string `json:"oldSize"`
	Price   string `json:"price"`
	OrderId string `json:"orderId"`
	Time    string `json:"time"`
}
