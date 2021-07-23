// +build active_file

package src

import (
	"container/list"
	"time"
)

type User struct {
	UserID      string
	CurrentPage *Page
	LastChange  time.Time
}

type Event struct {
	Timestamp int64   `json:"unix_timestamp"`
	PageTime  float64 `json:"page_time_seconds,omitempty"`
	Referrer  string  `json:"referrer,omitempty"`
	UserID    string  `json:"user_id"`
	Path      string  `json:"path"`
}

func (s *Simulator) Run() error {
	users := list.New()
	events := s.producer.TopicEncoder("events")

	for s.Running() {
		time.Sleep(JitterDuration(time.Second, 200*time.Millisecond))

		err := testTopic.Encode(map[string]interface{}{
			"message": "hello world",
			"time":    time.Now(),
			"worker":  s.id,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
