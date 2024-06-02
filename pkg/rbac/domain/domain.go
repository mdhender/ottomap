// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package domain

import (
	"encoding/json"
	"fmt"
)

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

// Role is a permission assigned to a user.
// If the user has the role, then the permission is granted.
type Role int

const (
	// Anonymous is a role that grants no permissions.
	Anonymous Role = iota
	// Administrator is a role that grants all permissions.
	Administrator
	// Authenticated is a role that grants basic permissions.
	Authenticated
	// Operator is a role that grants all permissions except for the administrator role.
	Operator
	// User is a role that grants basic permissions.
	User
)

// Contains returns true if the role is in the roles.
func (r Roles) Contains(role Role) bool {
	if r == nil {
		return role == Anonymous
	}
	_, ok := r[role]
	return ok
}

// EncodeText implements encoding.TextMarshaler and is used when encoding map keys to JSON.
func (r Role) EncodeText() ([]byte, error) {
	switch r {
	case Anonymous:
		return []byte("anonymous"), nil
	case Administrator:
		return []byte("administrator"), nil
	case Authenticated:
		return []byte("authenticated"), nil
	case Operator:
		return []byte("operator"), nil
	case User:
		return []byte("user"), nil
	}
	return nil, fmt.Errorf("invalid role %d", r)
}

func (r Role) MarshalJSON() ([]byte, error) {
	switch r {
	case Anonymous:
		return json.Marshal("anonymous")
	case Administrator:
		return json.Marshal("administrator")
	case Authenticated:
		return json.Marshal("authenticated")
	case Operator:
		return json.Marshal("operator")
	case User:
		return json.Marshal("user")
	}
	return nil, fmt.Errorf("invalid role %d", r)
}

func (r Role) String() string {
	switch r {
	case Anonymous:
		return "anonymous"
	case Administrator:
		return "administrator"
	case Authenticated:
		return "authenticated"
	case Operator:
		return "operator"
	case User:
		return "user"
	}
	panic(fmt.Sprintf("assert(role != %d)", r))
}

func (r *Role) UnmarshalJSON(b []byte) error {
	switch string(b) {
	case `"anonymous"`:
		*r = Anonymous
		return nil
	case `"administrator"`:
		*r = Administrator
		return nil
	case `"authenticated"`:
		*r = Authenticated
		return nil
	case `"operator"`:
		*r = Operator
		return nil
	case `"user"`:
		*r = User
		return nil
	}
	return fmt.Errorf("invalid role %s", string(b))
}
