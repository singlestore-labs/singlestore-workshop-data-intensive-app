// +build active_file

package src

import (
	"container/list"
	"math"
	"math/rand"
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
		unixNow := time.Now().Unix()

		// create a random number of new users
		if users.Len() < s.config.MaxUsersPerThread {

			// figure out the max number of users we can create
			maxNewUsers := s.config.MaxUsersPerThread - users.Len()

			// calculate a random number between 1 and maxNewUsers
			numNewUsers := RandomIntInRange(1, maxNewUsers)

			// create the new users
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

				// eventProb is a random value from 0 to 1 but is weighted
				// to be closer to 0 most of the time
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
