package src

import (
	"fmt"
	"log"
)

type Simulator struct {
	id      string
	config  *SimConfig
	closeCh chan struct{}
	stopped bool

	producer Producer
	sitemap  *Page
}

func NewSimulator(id string, config *SimConfig, sitemap *Page) (*Simulator, error) {
	p, err := NewFranzProducer(config.Producer)
	if err != nil {
		return nil, err
	}

	return &Simulator{
		id:       id,
		config:   config,
		closeCh:  make(chan struct{}),
		producer: p,
		sitemap:  sitemap,
	}, nil
}

func (s *Simulator) Running() bool {
	if !s.stopped {
		// If the closeCh is closed, we have been stopped.
		select {
		case <-s.closeCh:
			s.stopped = true
		default:
		}
	}
	return !s.stopped
}

func (s *Simulator) Stop() {
	s.logf("stopping")
	close(s.closeCh)
}

func (s *Simulator) logf(msg string, args ...interface{}) {
	log.Printf(fmt.Sprintf("[%s] %s", s.id, msg), args...)
}
