package events

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Kucoin/kucoin-level3-sdk/pkg/consts"
	"github.com/Kucoin/kucoin-level3-sdk/pkg/exchanges/kucoin-v2/sdk"
	"github.com/Kucoin/kucoin-level3-sdk/pkg/exchanges/kucoin-v2/stream"
	"github.com/Kucoin/kucoin-level3-sdk/pkg/services/log"
	"github.com/Kucoin/kucoin-level3-sdk/pkg/services/redis"
	"github.com/Kucoin/kucoin-level3-sdk/pkg/utils/orderbook/base"
	"go.uber.org/zap"
)

type OrderWatcher struct {
	Messages chan *sdk.WebSocketDownstreamMessage
	lock     *sync.RWMutex

	orderIds   map[string]map[string]bool //orderId => channel
	clientOids map[string]map[string]bool //clientOid => channel
}

func NewOrderWatcher() *OrderWatcher {
	return &OrderWatcher{
		Messages: make(chan *sdk.WebSocketDownstreamMessage, consts.MaxMsgChanLen),
		lock:     &sync.RWMutex{},

		orderIds:   make(map[string]map[string]bool),
		clientOids: make(map[string]map[string]bool),
	}
}

func (w *OrderWatcher) Run() {
	log.Info("start running OrderWatcher")

	for msg := range w.Messages {
		if !w.existEventOrderIds() {
			continue
		}

		l3Data, err := stream.NewStreamDataModel(msg)
		if err != nil {
			log.Panic("NewStreamDataModel err: " + err.Error())
			return
		}

		publishedData := base.ToJsonString(msg)
		switch l3Data.Type {
		case stream.MessageReceivedType:
			data := &stream.DataReceivedModel{}
			if err := json.Unmarshal(l3Data.Data(), data); err != nil {
				log.Panic("Unmarshal err", zap.Error(err))
			}

			w.migrationClientOidToOrderIds(data.ClientOid, data.OrderId)

			w.publish(data.OrderId, publishedData)

		case stream.MessageOpenType:
			data := &stream.DataOpenModel{}
			if err := json.Unmarshal(l3Data.Data(), data); err != nil {
				log.Panic("Unmarshal err", zap.Error(err))
			}

			w.publish(data.OrderId, publishedData)

		case stream.MessageMatchType:
			data := &stream.DataMatchModel{}
			if err := json.Unmarshal(l3Data.Data(), data); err != nil {
				log.Panic("Unmarshal err", zap.Error(err))
			}

			w.publish(data.MakerOrderId, publishedData)
			w.publish(data.TakerOrderId, publishedData)

		case stream.MessageDoneType:
			data := &stream.DataDoneModel{}
			if err := json.Unmarshal(l3Data.Data(), data); err != nil {
				log.Panic("Unmarshal err", zap.Error(err))
			}

			w.publish(data.OrderId, publishedData)
			w.removeEventOrderId(data.OrderId)

		case stream.MessageUpdateType:
			data := &stream.DataUpdateModel{}
			if err := json.Unmarshal(l3Data.Data(), data); err != nil {
				log.Panic("Unmarshal err", zap.Error(err))
			}

			w.publish(data.OrderId, publishedData)

		default:
			log.Panic("error msg type: " + l3Data.Type)
		}
	}
}

func (w *OrderWatcher) migrationClientOidToOrderIds(clientOid, orderId string) {
	w.lock.RLock()
	channelsMap, ok := w.clientOids[clientOid]
	var channels []string
	if ok {
		channels = getMapKeys(channelsMap)
	}
	w.lock.RUnlock()
	if ok {
		w.removeEventClientOid(clientOid)
		w.AddEventOrderIdsToChannels(map[string][]string{
			orderId: channels,
		})
	}
}

func getMapKeys(data map[string]bool) []string {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}

	return keys
}

func (w *OrderWatcher) publish(orderId string, message string) {
	w.lock.RLock()
	channelsMap, ok := w.orderIds[orderId]
	channels := getMapKeys(channelsMap)
	w.lock.RUnlock()

	if ok {
		for _, channel := range channels {
			redis.Publish("", channel, message)
		}
	}
}

func (w *OrderWatcher) existEventOrderIds() bool {
	w.lock.RLock()
	defer w.lock.RUnlock()
	if len(w.orderIds) == 0 && len(w.clientOids) == 0 {
		return false
	}

	return true
}

func (w *OrderWatcher) AddEventOrderIdsToChannels(data map[string][]string) {
	w.lock.Lock()
	defer w.lock.Unlock()

	log.Info(fmt.Sprintf("AddEventOrderIdsToChannels, %#v", data))
	for orderId, channels := range data {
		for _, channel := range channels {
			if w.orderIds[orderId] == nil {
				w.orderIds[orderId] = make(map[string]bool)
			}
			w.orderIds[orderId][channel] = true
		}
	}
}

func (w *OrderWatcher) AddEventClientOidsToChannels(data map[string][]string) error {
	w.lock.Lock()
	log.Info(fmt.Sprintf("AddEventClientOidsToChannels, %#v", data))
	for clientOid, channels := range data {
		for _, channel := range channels {
			if w.clientOids[clientOid] == nil {
				w.clientOids[clientOid] = make(map[string]bool)
			}
			w.clientOids[clientOid][channel] = true
		}
	}
	w.lock.Unlock()

	return nil
}

func (w *OrderWatcher) removeEventOrderId(orderId string) {
	w.lock.Lock()
	defer w.lock.Unlock()

	delete(w.orderIds, orderId)
}

func (w *OrderWatcher) removeEventClientOid(clientOid string) {
	w.lock.Lock()
	defer w.lock.Unlock()

	delete(w.clientOids, clientOid)
}
