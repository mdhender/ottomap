// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package users

import (
	"encoding/json"
	"os"
	"sync"
)

type Store struct {
	sync.RWMutex
	users map[string]*User
}

func New(path string) (*Store, error) {
	s := &Store{users: map[string]*User{}}

	users := map[string]*User{}
	if data, err := os.ReadFile(path); err != nil {
		return nil, err
	} else if err = json.Unmarshal(data, &users); err != nil {
		return nil, err
	}

	for id, user := range users {
		user.Id = id
		if user.Roles == nil {
			user.Roles = NewRoles()
		}
		user.IsAuthenticated = false
		s.users[id] = user
	}

	return s, nil
}

// MergeFrom loads the user store from the given path.
// It replaces users that are already in the store with the new values.
func (s *Store) MergeFrom(path string) error {
	s.Lock()
	defer s.Unlock()

	users := map[string]*User{}
	if data, err := os.ReadFile(path); err != nil {
		return err
	} else if err = json.Unmarshal(data, &users); err != nil {
		return err
	}

	for id, user := range users {
		if _, ok := s.users[id]; ok {
			delete(s.users, id)
		}
		user.Id = id
		if user.Roles == nil {
			user.Roles = NewRoles()
		}
		user.IsAuthenticated = false
		s.users[id] = user
	}

	return nil
}

func (s *Store) Authenticate(handle, secret string) (User, bool) {
	s.RLock()
	defer s.RUnlock()

	for _, user := range s.users {
		if user.Handle == handle && user.Secret == secret {
			cp := user.Clone()
			cp.IsAuthenticated = true
			return user.Clone(), true
		}
	}

	return User{}, false
}

func (s *Store) FetchById(id string) (User, bool) {
	s.RLock()
	defer s.RUnlock()

	if user, ok := s.users[id]; ok {
		return user.Clone(), true
	}

	return User{}, false
}

func (s *Store) TheSecrets() [][2]string {
	var da [][2]string
	for _, user := range s.users {
		da = append(da, [2]string{user.Handle, user.Secret})
	}
	return da
}
