// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package simba

type User struct {
	Id              string // unique identifier for the user
	Handle          string // unique handle (nickname) for the user
	Email           string // e-mail address for the user
	Secret          string // hashed secret for the user
	Clan            string // clan id
	IsAuthenticated bool   // derived from session
}
