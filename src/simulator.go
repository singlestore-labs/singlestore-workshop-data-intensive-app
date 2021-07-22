package src

import (
	"container/list"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"
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
		time.Sleep(time.Second)
		unixNow := time.Now().Unix()

		// create a random number of new users
		numNewUsers := int(math.Max(1, float64(s.config.MaxUsersPerTick)*rand.Float64()))
		for i := 0; i < numNewUsers; i++ {
			// define a new user
			user := &User{
				UserID:      NextUserId(),
				CurrentPage: s.sitemap.RandomLeaf(),
				LastChange:  time.Now(),
			}

			// add the user to the list
			users.PushBack(user)

			// write an event to the topic
			err := events.Encode(Event{
				Timestamp: unixNow,
				UserID:    user.UserID,
				Path:      user.CurrentPage.Path,

				// we provide a fake referrer here
				Referrer: RandomReferrer(),
			})
			if err != nil {
				return err
			}
		}

		// we will be removing elements from the list while we iterate, so we
		// need to keep track of next outside of the loop
		var next *list.Element

		// iterate through the users list and simulate each users behavior
		for el := users.Front(); el != nil; el = next {
			next = el.Next()
			user := el.Value.(*User)
			pageTime := time.Since(user.LastChange)

			// users only consider leaving a page after at least 5 seconds
			if pageTime > time.Second*5 {
				eventProb := math.Pow(rand.Float64(), 2)
				if eventProb > 0.98 {
					// user has left the site
					users.Remove(el)
					continue
				} else if eventProb > 0.9 {
					// user jumps to a random page
					user.CurrentPage = s.sitemap.RandomLeaf()
					user.LastChange = time.Now()
				} else if eventProb > 0.8 {
					// user goes to the "next" page
					user.CurrentPage = user.CurrentPage.RandomNext()
					user.LastChange = time.Now()
				}
			}

			err := events.Encode(Event{
				Timestamp: unixNow,
				UserID:    user.UserID,
				Path:      user.CurrentPage.Path,
				PageTime:  pageTime.Seconds(),
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
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
