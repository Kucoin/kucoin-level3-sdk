package verify

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Kucoin/kucoin-level3-sdk/pkg/exchanges/kucoin-v2/orderbook"
	"github.com/Kucoin/kucoin-level3-sdk/pkg/exchanges/kucoin-v2/sdk"
	"github.com/Kucoin/kucoin-level3-sdk/pkg/services/log"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type Verify struct {
	level3Builder   *orderbook.Builder
	Messages        chan *sdk.WebSocketDownstreamMessage //方便记录两次快照之间的数据流，方便在快照校验失败后，重复快照
	verifyFrequency time.Duration                        //校验频率

	verifyLogDirectory string
	update             *os.File

	uniqStr string

	nextVerifyTime      time.Time
	snapshot            map[uint64]*orderbook.FullOrderBook
	atomicFullOrderBook *orderbook.DepthResponse
}

func NewVerify(level3Builder *orderbook.Builder, verifyFrequency uint64, verifyLogDirectory string, uniqStr string) *Verify {
	verifyLogDirectory = strings.TrimRight(verifyLogDirectory, "/")
	v := &Verify{
		level3Builder:      level3Builder,
		Messages:           make(chan *sdk.WebSocketDownstreamMessage, 1024),
		verifyFrequency:    time.Duration(verifyFrequency) * time.Second,
		verifyLogDirectory: strings.TrimRight(verifyLogDirectory, "/") + "/",
		uniqStr:            uniqStr,
	}

	v.resetVerify(true)

	//第一次校验时间在启动后2分钟开始开启
	v.nextVerifyTime = time.Now().Add(10 * time.Second)

	return v
}

//重置数据
func (v *Verify) resetVerify(setNextVerifyTime bool) {
	if setNextVerifyTime {
		v.nextVerifyTime = time.Now().Add(v.verifyFrequency)
	} else { //防止请求过多
		v.nextVerifyTime = time.Now().Add(time.Minute)
	}
	v.snapshot = map[uint64]*orderbook.FullOrderBook{}
	v.atomicFullOrderBook = nil
}

//func (v *Verify) Log() {
//	for msg := range v.Messages {
//		log.Debug("verify websocket", zap.Any("type", msg.Subject), zap.Any("Data", string(msg.RawData)))
//	}
//}

func (v *Verify) Run() {
	defer v.update.Close()

	v.checkLogDirExists()

	log.Info("start running Verify, verifyLogDirectory: "+v.verifyLogDirectory, zap.Any("verifyFrequency", v.verifyFrequency))

	const snapshotMaxLen = 200
	const atomicStartLen = 10

	for msg := range v.Messages {
		v.writeWsMsg(msg)
		if time.Now().After(v.nextVerifyTime) {
			//获取快照
			snapshot, err := v.level3Builder.Snapshot()
			if err != nil {
				panic(err)
			}

			if snapshot.Sequence == 0 {
				continue
			}

			if _, ok := v.snapshot[snapshot.Sequence]; ok {
				//跳过已经存在的快照
				continue
			}
			v.snapshot[snapshot.Sequence] = snapshot

			//log.Error("准备校验%d", len(v.snapshot))

			//获取全量
			if len(v.snapshot) == atomicStartLen {
				//log.Error("获取全量...")

				go func() {
					var err error
					log.Info("verify 获取全量")

					v.atomicFullOrderBook, err = v.level3Builder.GetAtomicFullOrderBook()
					if err != nil {
						log.Warn("获取全量失败放弃本次校验", zap.Error(err))
						v.resetVerify(false)
					}
				}()
			}

			if v.atomicFullOrderBook != nil {
				//开始校验 !!!
				atomicFullOrderBookSequence := v.atomicFullOrderBook.Sequence
				if snapshot, ok := v.snapshot[atomicFullOrderBookSequence]; ok {
					log.Info("verify 开始校验", zap.Any("Sequence", snapshot.Sequence))

					//写入全量快照
					if v.verifyLogDirectory != "" {
						data, _ := json.Marshal(snapshot)
						v.writeSnapshot(fmt.Sprintf("%d.snapshot.json", snapshot.Sequence), data)
					}

					sortedAtomicFullOrderBook, err := v.level3Builder.DepthResponse2FullOrderBook(v.atomicFullOrderBook)
					if err != nil {
						log.Panic(fmt.Sprintf("verify order book快照校验失败, DepthResponse2FullOrderBook 转换失败, Sequence: %d, error: %s", snapshot.Sequence, err.Error()))
					}
					if err := v.diffOrderBook(snapshot, sortedAtomicFullOrderBook); err != nil {
						if v.verifyLogDirectory != "" {
							data, _ := json.Marshal(sortedAtomicFullOrderBook)
							v.writeSnapshot(fmt.Sprintf("%d.atomicFullOrderBook.json", v.atomicFullOrderBook.Sequence), data)
						}
						log.Panic(fmt.Sprintf("verify order book快照校验失败, Sequence: %d, error: %s", snapshot.Sequence, err.Error()))
						v.resetVerify(true)
						continue
					} else {
						//开启一次新的消息流
						v.getNewFile(fmt.Sprintf("%d.log", snapshot.Sequence))
					}

					log.Info("verify Success !!! order book 快照比对成功", zap.Any("Sequence", snapshot.Sequence))
					v.resetVerify(true)
					continue
				}

				//以下逻辑提高校验效率，去掉无用的全量和后续增量获取
				if atomicFullOrderBookSequence > snapshot.Sequence+uint64(snapshotMaxLen)-uint64(len(v.snapshot)) {
					log.Warn("获取到太大的全量", zap.Any("full", atomicFullOrderBookSequence), zap.Any("curr", snapshot.Sequence), zap.Any("len", len(v.snapshot)))
					v.resetVerify(false)
					continue
				}

				if atomicFullOrderBookSequence < snapshot.Sequence-uint64(len(v.snapshot)) {
					log.Warn("获取到太小的全量", zap.Any("full", atomicFullOrderBookSequence), zap.Any("curr", snapshot.Sequence), zap.Any("len", len(v.snapshot)))
					v.resetVerify(false)
					continue
				}
			}

			//最多20份快照
			if len(v.snapshot) > snapshotMaxLen {
				log.Warn("跳过校验...")
				v.resetVerify(false)
			}
		}
	}
}

func (v *Verify) diffOrderBook(snapshot, atomicFullOrderBook *orderbook.FullOrderBook) error {
	if snapshot.Sequence != atomicFullOrderBook.Sequence {
		return errors.New(fmt.Sprintf("sequence 不一致: %d - %d", snapshot.Sequence, atomicFullOrderBook.Sequence))
	}

	if err := v.diffOrderBookOneSide(snapshot.Asks, atomicFullOrderBook.Asks); err != nil {
		return errors.New("asks 比对不一致, " + err.Error())
	}

	if err := v.diffOrderBookOneSide(snapshot.Bids, atomicFullOrderBook.Bids); err != nil {
		return errors.New("bids 比对不一致, " + err.Error())
	}

	return nil
}

func diff(a string, b string) error {
	if a == b {
		return nil
	}

	aF, err := decimal.NewFromString(a)
	if err != nil {
		return err
	}
	bF, err := decimal.NewFromString(b)
	if err != nil {
		return err
	}

	if !aF.Equal(bF) {
		return errors.New("不相等: " + a + " != " + b)
	}

	return nil
}

func (v *Verify) diffOrderBookOneSide(items, atomicItems [][3]string) error {
	if len(items) != len(atomicItems) {
		return errors.New("比对失败，深度不一样")
	}

	for index, item := range items {
		//[3]string{"orderId", "price", "size"}
		if item[0] != atomicItems[index][0] {
			return errors.New(fmt.Sprintf("比对失败, index: %d, 订单id: %s != %s", index, item[0], atomicItems[index][0]))
		}

		if err := diff(item[1], atomicItems[index][1]); err != nil {
			return errors.New(fmt.Sprintf("price 比对失败, order id: %s, error: %s", item[0], err.Error()))
		}

		if err := diff(item[2], atomicItems[index][2]); err != nil {
			return errors.New(fmt.Sprintf("size 比对失败, order id: %s, error: %s", item[0], err.Error()))
		}
	}

	return nil
}

func (v *Verify) getNewFile(filename string) {
	//先关闭之前
	if v.update != nil {
		_ = v.update.Close()
	}

	filename = v.verifyLogDirectory + v.uniqStr + "-update-" + filename
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}

	v.update = f
}

func (v *Verify) writeWsMsg(data *sdk.WebSocketDownstreamMessage) {
	if v.verifyLogDirectory == "" {
		return
	}

	if v.update == nil {
		//刚开始未知seq，初始化一个 file
		v.getNewFile(fmt.Sprintf("%s.log", time.Now().Format("2006-01-02-15-04-05")))
	}
	msg, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	msg = append(msg, '\n')

	if _, err := v.update.Write(msg); err != nil {
		panic(err)
	}
}

func (v *Verify) writeSnapshot(filename string, data []byte) {
	filename = v.verifyLogDirectory + v.uniqStr + "-" + filename
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		panic(err)
	}
}

func (v *Verify) checkLogDirExists() {
	if v.verifyLogDirectory == "/" || v.verifyLogDirectory == "" {
		v.verifyLogDirectory = ""
		log.Warn("日志目录为空，不记录日志")
		return
	}

	if _, err := os.Stat(v.verifyLogDirectory); os.IsNotExist(err) {
		panic("日志目录不存在: " + v.verifyLogDirectory)
	}
}
