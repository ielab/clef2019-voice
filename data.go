package main

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
)

type User struct {
	Name   string   `json:"string"`
	Topics []string `json:"topics"`
}

func (s server) markTopicAsCompleted(user string, topic Topic) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Get the study bucket.
		b, err := tx.CreateBucketIfNotExists([]byte("study"))
		if err != nil {
			return err
		}

		// Get the user bucket in the study bucket.
		ub := b.Bucket([]byte(user))
		if ub == nil {
			return fmt.Errorf("user does not exist")
		}

		// Get the topic and mark it as completed.
		topic.Completed = true
		v, err := json.Marshal(topic)
		if err != nil {
			return err
		}

		return ub.Put([]byte(topic.Topic), v)
	})
}

func (s server) getTopics(user string) ([]Topic, error) {
	var ts []Topic
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("study"))

		if b == nil {
			return fmt.Errorf("study table does not exist (add a user)")
		}

		ub := b.Bucket([]byte(user))
		if ub == nil {
			return fmt.Errorf("user does not exist")
		}

		return ub.ForEach(func(k, v []byte) error {
			var t Topic
			err := json.Unmarshal(v, &t)
			if err != nil {
				return err
			}
			ts = append(ts, t)
			return nil
		})
	})
	return ts, err
}

func (s server) getTopic(user, topic string) (Topic, error) {
	var t Topic
	err := s.db.View(func(tx *bolt.Tx) error {
		// Get the study bucket.
		b := tx.Bucket([]byte("study"))
		if b == nil {
			return fmt.Errorf("study bucket not initialised")
		}

		// Get the user bucket in the study bucket.
		ub := b.Bucket([]byte(user))
		if ub == nil {
			return fmt.Errorf("user does not exist")
		}

		// Get the topic specified.
		v := ub.Get([]byte(topic))
		if v == nil {
			return fmt.Errorf("user is not assigned that topic")
		}

		return json.Unmarshal(v, &t)
	})
	return t, err
}

func (s server) updateTopic(user string, topic Topic) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("study"))

		ub := b.Bucket([]byte(user))

		v, err := json.Marshal(topic)
		if err != nil {
			return err
		}

		return ub.Put([]byte(topic.Topic), v)
	})
}

func (s server) getUsers() ([]User, error) {
	var u []User
	err := s.db.Update(func(tx *bolt.Tx) error {
		// Get the study bucket.
		b, err := tx.CreateBucketIfNotExists([]byte("admin"))
		if err != nil {
			return err
		}

		return b.ForEach(func(k, v []byte) error {
			var user User
			err := json.Unmarshal(v, &user)
			if err != nil {
				return err
			}
			u = append(u, user)
			return nil
		})
	})
	return u, err
}

func (s server) addUser(user string, topics []string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Get the study bucket.
		b, err := tx.CreateBucketIfNotExists([]byte("admin"))
		if err != nil {
			return err
		}

		// Get the study bucket.
		s, err := tx.CreateBucketIfNotExists([]byte("study"))
		if err != nil {
			return err
		}

		// Get the user bucket in the study bucket.
		us, err := s.CreateBucketIfNotExists([]byte(user))
		if err != nil {
			return err
		}

		// Add topics to the user in their bucket.
		for _, topic := range topics {
			t := Topic{
				User:      user,
				Topic:     topic,
				Completed: false,
			}
			v, err := json.Marshal(t)
			if err != nil {
				return err
			}
			err = us.Put([]byte(topic), v)
			if err != nil {
				return err
			}
		}

		// Add the user to the admin page.
		v, err := json.Marshal(User{Name: user, Topics: topics})
		if err != nil {
			return err
		}
		return b.Put([]byte(user), v)
	})
}

func (s server) removeUser(user string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Get the study bucket.
		b := tx.Bucket([]byte("admin"))
		if b == nil {
			return fmt.Errorf("admin bucket does not exist")
		}
		err := b.Delete([]byte(user))
		if err != nil {
			return err
		}

		// Get the study bucket.
		s := tx.Bucket([]byte("study"))
		if s == nil {
			return fmt.Errorf("study bucket does not exist")
		}
		return s.DeleteBucket([]byte(user))
	})
}
