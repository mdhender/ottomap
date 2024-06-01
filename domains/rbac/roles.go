// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package rbac

import "sync"

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

// Store is a store for roles.
// It is the link between the roles and the users.
type Store struct {
	sync.RWMutex

	// the key to the roles is the user id.
	// the value is the role assigned to the user.
	roles map[string]Roles

	anonymousUser Roles
}

// NewStore returns a new store.
func NewStore() *Store {
	return &Store{
		roles:         map[string]Roles{},
		anonymousUser: NewRoles("anonymous"),
	}
}

func (s *Store) AnonymousUser() Roles {
	return s.anonymousUser
}

// FetchRolesByUserId returns the roles assigned to the user with the given id.
func (s *Store) FetchRolesByUserId(id string) (Roles, bool) {
	s.RLock()
	defer s.RUnlock()

	roles, ok := s.roles[id]
	if !ok {
		return map[Role]bool{}, false
	}
	return roles, ok
}

func (s *Store) UserRoles(id string) Roles {
	roles, ok := s.FetchRolesByUserId(id)
	if !ok {
		return s.AnonymousUser()
	}
	return roles
}

// UserHasRole returns true if the user has the given role.
// Returns false if the user does not exist or does not have the role.
func (s *Store) UserHasRole(id string, role Role) bool {
	roles, ok := s.FetchRolesByUserId(id)
	if !ok {
		return false
	}
	return roles[role]
}

// SetUserRoles replaces the roles assigned to the user with the given id.
func (s *Store) SetUserRoles(id string, roles Roles) {
	s.Lock()
	defer s.Unlock()
	s.roles[id] = roles.Clone()
}
