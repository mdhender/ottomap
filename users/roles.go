// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package users

// Roles is a map of the permissions assigned to a user
type Roles map[string]bool

// NewRoles returns a new roles map with the assigned roles.
func NewRoles(roles ...string) Roles {
	r := map[string]bool{}
	for _, role := range roles {
		r[role] = true
	}
	return r
}

// Clone returns a deep copy of the roles.
func (r Roles) Clone() Roles {
	cp := NewRoles()
	for k, v := range r {
		if v {
			cp[k] = v
		}
	}
	return cp
}
