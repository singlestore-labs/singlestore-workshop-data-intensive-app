package src

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/pkg/errors"
	"github.com/twmb/franz-go/pkg/kgo"
)

type Producer interface {
	TopicEncoder(topic string) *json.Encoder
	Close() error
}

type FranzProducer struct {
	client        *kgo.Client
	closed        int32 // nonzero if the producer has started closing. accessed via atomics
	pendingWrites sync.WaitGroup
}

type ProducerConfig struct {
	Brokers []string
}

func NewFranzProducer(config ProducerConfig) (Producer, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(config.Brokers...),
		kgo.MaxBufferedRecords(1e7),
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	return &FranzProducer{
		client: client,
	}, nil
}

func (p *FranzProducer) TopicEncoder(topic string) *json.Encoder {
	return json.NewEncoder(&FranzWriter{p: p, topic: topic})
}

func (p *FranzProducer) Closed() bool {
	return atomic.LoadInt32(&p.closed) != 0
}

func (p *FranzProducer) Close() error {
	if p.Closed() {
		return errors.New("already closed")
	}
	atomic.StoreInt32(&p.closed, 1)

	p.pendingWrites.Wait()

	p.client.Close()

	return nil
}

type FranzWriter struct {
	p     *FranzProducer
	topic string
}

func (w *FranzWriter) Write(d []byte) (int, error) {
	if w.p.Closed() {
		return 0, syscall.EINVAL
	}

	r := kgo.SliceRecord(d)
	r.Topic = w.topic

	w.p.pendingWrites.Add(1)
	w.p.client.Produce(context.Background(), r, func(r *kgo.Record, err error) {
		defer w.p.pendingWrites.Done()
		if err != nil && err != kgo.ErrClientClosed {
			log.Printf("FranzProducer error on topic %s: %+v", w.topic, err)
		}
	})

	return len(d), nil
}
