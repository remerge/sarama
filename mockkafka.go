package dm

// TODO MOVE TO REX

import (
	"testing"
	"time"

	"github.com/Shopify/sarama"
)

func Terminated(ms time.Duration, f func(), terminator func()) bool {
	ok := make(chan bool)
	if terminator != nil {
		time.AfterFunc(10*time.Millisecond, terminator)
	}
	go func() {
		f()
		ok <- true
	}()
	select {
	case <-ok:
		return true
	case <-time.After(ms * time.Millisecond):
		return false
	}
}

type MockKafka struct {
	mb1             *sarama.MockBroker
	mb2             *sarama.MockBroker
	offset          int64
	topic           string
	t               *testing.T
	requestHandlers map[string]sarama.MockResponse
}

func InitMockKafka(topic string, t *testing.T) *MockKafka {
	k := &MockKafka{
		topic: topic,
		t:     t,
	}

	k.mb1 = sarama.NewMockBroker(t, 1)
	k.mb2 = sarama.NewMockBroker(t, 2)
	k.requestHandlers = make(map[string]sarama.MockResponse)
	k.AddMetadataResponse()

	return k
}

// MockOffsetResponse mocks offset response
func (k *MockKafka) MockOffsetResponse(topic string, partition int32, offset int64) {
	offsetHandler := k.requestHandlers["OffsetRequest"]
	if offsetHandler == nil {
		offsetHandler = sarama.NewMockOffsetResponse(k.t)
	}
	offsetHandler = offsetHandler.(*sarama.MockOffsetResponse).
		SetOffset(topic, partition, sarama.OffsetOldest, offset).
		SetOffset(topic, partition, sarama.OffsetNewest, offset)
	k.requestHandlers["OffsetRequest"] = offsetHandler

	k.mb2.SetHandlerByMap(k.requestHandlers)
}

func (k *MockKafka) AddMetadataResponse() {
	mdr := new(sarama.MetadataResponse)
	mdr.AddBroker(k.mb2.Addr(), k.mb2.BrokerID())
	mdr.AddTopicPartition(k.topic, 0, k.mb2.BrokerID(), nil, nil, sarama.ErrNoError)
	k.mb1.Returns(mdr)
}

func (k *MockKafka) Addr() string {
	return k.mb1.Addr()
}

func MockMsgAddId(id []byte, topic string, partition int32, offset int64, b *sarama.MockBroker) {
	msg := make([]byte, 21)
	copy(msg[1:], id)
	msg[0] = 1 // insert
	MockMsg(msg, topic, partition, offset, b)
}

func MockMsg(msg []byte, topic string, partition int32, offset int64, b *sarama.MockBroker) {
	fr := new(sarama.FetchResponse)
	fr.AddMessage(topic, partition, nil, sarama.ByteEncoder(msg), offset)
	b.Returns(fr)
}

func (k *MockKafka) Msg(msg []byte) {
	fr := new(sarama.FetchResponse)
	fr.AddMessage(k.topic, 0, nil, sarama.ByteEncoder(msg), k.offset)
	k.offset++
	k.mb2.Returns(fr)
}

func (k *MockKafka) MsgStr(msg string) {
	fr := new(sarama.FetchResponse)
	fr.AddMessage(k.topic, 0, nil, sarama.StringEncoder(msg), k.offset)
	k.offset++
	k.mb2.Returns(fr)
}

func (k *MockKafka) AddId(id []byte) {
	msg := make([]byte, 21)
	copy(msg[1:], id)
	msg[0] = 1 // insert
	k.Msg(msg)
}

func (k *MockKafka) RemoveId(id []byte) {
	msg := make([]byte, 21)
	copy(msg[1:], id)
	msg[0] = 2 // remove
	k.Msg(msg)
}
