// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package domain

import (
	users "github.com/mdhender/ottomap/pkg/users/domain"
	"time"
)

// Session is the data for a current session
type Session struct {
	Id        string
	ExpiresAt time.Time
	User      users.User
}

// IsExpired returns true if the session is expired.
func (s *Session) IsExpired() bool {
	return !time.Now().Before(s.ExpiresAt)
}
