// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package dao

import (
	rbac "github.com/mdhender/ottomap/pkg/rbac/domain"
	"sync"
)

// Store is a store for roles.
// It is the link between the roles and the users.
type Store struct {
	sync.RWMutex

	// the key to the roles is the user id.
	// the value is the role assigned to the user.
	roles map[string]rbac.Roles

	anonymousUser rbac.Roles
}

// NewStore returns a new store.
func NewStore() *Store {
	return &Store{
		roles:         map[string]rbac.Roles{},
		anonymousUser: rbac.NewRoles(rbac.Anonymous),
	}
}

func (s *Store) AnonymousUser() rbac.Roles {
	return s.anonymousUser
}

// GetUserRoles returns the roles assigned to the user with the given id.
func (s *Store) GetUserRoles(id string) (rbac.Roles, bool) {
	s.RLock()
	defer s.RUnlock()

	roles, ok := s.roles[id]
	if !ok {
		return map[rbac.Role]bool{}, false
	}
	return roles, ok
}

// UserHasRole returns true if the user has the given role.
// Returns false if the user does not have the role.
func (s *Store) UserHasRole(id string, role rbac.Role) bool {
	roles, ok := s.GetUserRoles(id)
	if !ok {
		return false
	}
	return roles.Contains(role)
}

// SetUserRoles replaces the roles assigned to the user with the given id.
func (s *Store) SetUserRoles(id string, roles ...rbac.Role) {
	s.Lock()
	defer s.Unlock()
	s.roles[id] = rbac.NewRoles(roles...)
}
