// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package users

type User struct {
	Id     string `json:"id"`               // unique identifier for the user
	Handle string `json:"handle,omitempty"` // unique handle (nickname) for the user
	Email  string `json:"email,omitempty"`  // e-mail address for the user
	Secret string `json:"secret,omitempty"` // hashed secret for the user
	Roles  Roles  `json:"roles,omitempty"`

	// helper values that don't get saved to the store
	Clan            string `json:"-"` // clan id
	IsAuthenticated bool   `json:"-"`
}

func (u User) Clone() User {
	return User{
		Id:     u.Id,
		Handle: u.Handle,
		Secret: u.Secret,
		Roles:  u.Roles.Clone(),
	}
}
