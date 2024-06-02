// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package simba

import (
	"time"
)

type Session struct {
	id        string
	expiresAt time.Time
	user      *User
	agent     *Agent
}
