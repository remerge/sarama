package sarama

import (
	"fmt"
)

// TopicConsumer consumes a single topic starting from a given offset.
type TopicConsumer interface {
	// Messages returns the read channel for the messages that are returned by
	// the broker.
	Messages() <-chan *ConsumerMessage

	Errors() <-chan error

	Close() error
}

type consumedPartition struct {
	forwarder         *forwarder
	partitionConsumer PartitionConsumer
}

type topicConsumer struct {
	topic              string
	client             Client
	master             Consumer
	consumedPartitions []consumedPartition
	messages           chan *ConsumerMessage
	errors             chan error
	offsetCorrection   bool
}

func NewTopicConsumer(client Client, topic string, offsets map[int32]int64, offsetCorrection bool) (TopicConsumer, error) {
	partitions, err := client.Partitions(topic)

	if err != nil {
		return nil, err
	}

	master, err := NewConsumerFromClient(client)

	if err != nil {
		return nil, err
	}

	consumer := &topicConsumer{
		topic:            topic,
		client:           client,
		master:           master,
		messages:         make(chan *ConsumerMessage, client.Config().ChannelBufferSize),
		errors:           make(chan error),
		offsetCorrection: offsetCorrection,
	}

	for _, partition := range partitions {
		err = consumer.initPartition(partition, offsets[partition])
		if err != nil {
			return nil, fmt.Errorf("Unable to init partition  %v: %v", partition, err)
		}
	}

	return consumer, nil
}

func (sc *topicConsumer) initPartition(partition int32, offset int64) error {
	oldestOffset, err := sc.client.GetOffset(sc.topic, partition, OffsetOldest)

	if err != nil {
		return err
	}

	newestOffset, err := sc.client.GetOffset(sc.topic, partition, OffsetNewest)

	if err != nil {
		return err
	}

	resumeFrom := offset

	if !sc.offsetCorrection && (oldestOffset > resumeFrom || newestOffset < resumeFrom) {
		return fmt.Errorf("offset for %v/%v is out of range of available offsets (%v..%v)", sc.topic, partition, newestOffset, oldestOffset)
	}

	if oldestOffset > resumeFrom {
		Logger.Printf("given offset for %v/%v is unavailable (oldest > resume_from). Auto correcting to oldest available offset: resume_from=%v oldest=%v", sc.topic, partition, resumeFrom, oldestOffset)
		resumeFrom = oldestOffset
	}

	if newestOffset < resumeFrom {
		Logger.Printf("given offset for %v/%v is unavailable (newest < resume_from). Auto correcting to oldest available offset: resume_from=%v newest=%v", sc.topic, partition, resumeFrom, newestOffset)
		resumeFrom = newestOffset
	}

	partitionConsumer, err := sc.master.ConsumePartition(sc.topic, partition, resumeFrom)

	if err != nil {
		return err
	}

	forwarder := newForwarder(partitionConsumer.Messages(), partitionConsumer.Errors())

	go forwarder.forwardTo(sc.messages, sc.errors)

	sc.consumedPartitions = append(sc.consumedPartitions, consumedPartition{
		forwarder:         forwarder,
		partitionConsumer: partitionConsumer,
	})

	return nil
}

func (sc *topicConsumer) Messages() <-chan *ConsumerMessage {
	return sc.messages
}

func (sc *topicConsumer) Errors() <-chan error {
	return sc.errors
}

func (sc *topicConsumer) Close() error {
	for _, consumedPartition := range sc.consumedPartitions {
		consumedPartition.forwarder.Close()
		consumedPartition.partitionConsumer.Close()
	}

	if err := sc.master.Close(); err != nil {
		return err
	}

	close(sc.messages)

	return nil
}
