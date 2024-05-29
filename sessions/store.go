// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package sessions

import (
	"encoding/json"
	"github.com/mdhender/ottomap/users"
	"os"
	"sync"
	"time"
)

// Store is a store for sessions.
type Store struct {
	sync.RWMutex
	sessions map[string]*Session
}

type Session struct {
	Id        string     `json:"-"`
	ExpiresAt time.Time  `json:"expiresAt"`
	User      users.User `json:"user"`
}

func NewStore(path string, us *users.Store) (*Store, error) {
	s := &Store{sessions: map[string]*Session{}}

	sessions := map[string]*Session{}
	if data, err := os.ReadFile(path); err != nil {
		return nil, err
	} else if err = json.Unmarshal(data, &sessions); err != nil {
		return nil, err
	}

	for id, sess := range sessions {
		sess.Id = id

		user, ok := us.FetchById(sess.User.Id)
		if !ok { // ignore sessions for users that no longer exist
			continue
		}
		user.IsAuthenticated = true
		s.sessions[id] = &Session{
			Id:        id,
			ExpiresAt: time.Now().Add(2 * 7 * 24 * time.Hour), // 2 weeks,
			User:      user,
		}
	}

	return s, nil
}

func (ss *Store) MergeFrom(path string, us *users.Store) error {
	ss.Lock()
	defer ss.Unlock()

	sessions := map[string]*Session{}
	if data, err := os.ReadFile(path); err != nil {
		return err
	} else if err = json.Unmarshal(data, &sessions); err != nil {
		return err
	}

	for id, sess := range sessions {
		if _, ok := ss.sessions[id]; ok {
			delete(ss.sessions, id)
		}
		user, ok := us.FetchById(sess.User.Id)
		if !ok { // ignore sessions for users that no longer exist
			continue
		}
		user.IsAuthenticated = true
		ss.sessions[id] = &Session{
			Id:        id,
			ExpiresAt: time.Now().Add(2 * 7 * 24 * time.Hour), // 2 weeks,
			User:      user,
		}
	}

	return nil
}

func (ss *Store) add(sess *Session) {
	ss.Lock()
	defer ss.Unlock()

	if _, ok := ss.sessions[sess.Id]; ok {
		delete(ss.sessions, sess.Id)
	}

	ss.sessions[sess.Id] = sess
}

func (ss *Store) del(id string) {
	ss.Lock()
	defer ss.Unlock()

	delete(ss.sessions, id)
}

func (ss *Store) get(id string) (*Session, bool) {
	ss.RLock()
	defer ss.RUnlock()

	sess, ok := ss.sessions[id]
	if !ok {
		return nil, false
	} else if !time.Now().Before(sess.ExpiresAt) {
		delete(ss.sessions, id)
	}

	return sess, ok
}
