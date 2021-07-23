// +build active_file

package src

import (
	"time"
)

func (s *Simulator) Run() error {
	testTopic := s.producer.TopicEncoder("test")

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
