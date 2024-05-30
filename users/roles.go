// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package users

// Role is a permission assigned to a user.
// If the user has the role, then the permission is granted.
type Role string

// Roles is a map of the permissions assigned to a user.
type Roles map[Role]bool

// NewRoles returns a new roles map with the assigned roles.
func NewRoles(roles ...Role) Roles {
	r := map[Role]bool{}
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
