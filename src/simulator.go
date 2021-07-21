package src

import (
	"fmt"
	"log"
	"time"
)

type Simulator struct {
	id      string
	config  *SimConfig
	closeCh chan struct{}

	producer Producer
}

func NewSimulator(id string, config *SimConfig) (*Simulator, error) {
	p, err := NewFranzProducer(config.Producer)
	if err != nil {
		return nil, err
	}

	return &Simulator{
		id:       id,
		config:   config,
		closeCh:  make(chan struct{}),
		producer: p,
	}, nil
}

func (s *Simulator) Run() error {
	for {
		time.Sleep(time.Second)

		s.logf("doing some work...")

		testTopic := s.producer.TopicEncoder("test")
		testTopic.Encode(map[string]interface{}{
			"worker":  s.id,
			"message": "hello world",
			"time":    time.Now(),
		})

		// If the closeCh is closed, we have been stopped.
		select {
		case <-s.closeCh:
			s.logf("exited")
			return nil
		default:
		}
	}
}

func (s *Simulator) Stop() {
	s.logf("stopping")
	close(s.closeCh)
}

func (s *Simulator) logf(msg string, args ...interface{}) {
	log.Printf(fmt.Sprintf("[%s] %s", s.id, msg), args...)
}
