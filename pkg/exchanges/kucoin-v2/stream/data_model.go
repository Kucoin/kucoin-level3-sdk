package stream

import (
	"encoding/json"

	"github.com/Kucoin/kucoin-level3-sdk/pkg/exchanges/kucoin-v2/sdk"
)

type SequenceModel struct {
	Sequence uint64 `json:"sequence"`
}

//Level 3 websocket stream
type DataModel struct {
	Type string `json:"type"`

	Sequence uint64 `json:"sequence"`

	//Symbol     string `json:"symbol"`

	rawData json.RawMessage
}

func NewStreamDataModel(msgData *sdk.WebSocketDownstreamMessage) (*DataModel, error) {
	data := &SequenceModel{}
	if err := json.Unmarshal(msgData.RawData, data); err != nil {
		return nil, err
	}

	return &DataModel{
		Type:     msgData.Subject,
		Sequence: data.Sequence,
		rawData:  msgData.RawData,
	}, nil
}

func (l3Data *DataModel) Data() json.RawMessage {
	return l3Data.rawData
}

const (
	BuySide  = "buy"
	SellSide = "sell"

	MessageReceivedType = "received"
	MessageOpenType     = "open"
	MessageDoneType     = "done"
	MessageMatchType    = "match"
	MessageUpdateType   = "update"
)

type DataReceivedModel struct {
	OrderId   string `json:"orderId"`
	ClientOid string `json:"clientOid"`
	Time      uint64 `json:"ts"`
}

type DataOpenModel struct {
	Side      string `json:"side"`
	Size      string `json:"size"`
	Price     string `json:"price"`
	OrderId   string `json:"orderId"`
	OrderTime uint64 `json:"orderTime"`
	Time      uint64 `json:"ts"`
}

type DataUpdateModel struct {
	OrderId string `json:"orderId"`
	Size    string `json:"size"`
	Time    uint64 `json:"ts"`
}

type DataMatchModel struct {
	Side         string `json:"side"`
	Price        string `json:"price"`
	Size         string `json:"size"`
	RemainSize   string `json:"remainSize"`
	TakerOrderId string `json:"takerOrderId"`
	MakerOrderId string `json:"makerOrderId"`
	TradeId      string `json:"tradeId"`
	Time         uint64 `json:"ts"`
}

type DataDoneModel struct {
	OrderId string `json:"orderId"`
	Reason  string `json:"reason"`
	Time    uint64 `json:"ts"`
}
