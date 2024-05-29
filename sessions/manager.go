// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package sessions

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/mdhender/ottomap/users"
	"log"
	"net/http"
	"sync"
	"time"
)

// Manager manages the sessions for the application.
type Manager struct {
	sync.RWMutex
	cookieName string
	sessions   *Store
	users      *users.Store
}

func NewManager(cookieName string, sessions *Store, us *users.Store) (*Manager, error) {
	if len(cookieName) == 0 {
		return nil, fmt.Errorf("bad cookie")
	} else if sessions == nil {
		return nil, fmt.Errorf("missing sessions store")
	} else if us == nil {
		return nil, fmt.Errorf("missing users store")
	}
	return &Manager{
		cookieName: cookieName + "-session",
		sessions:   sessions,
		users:      us,
	}, nil
}

func (sm *Manager) CreateSession(userId string) (string, bool) {
	user, ok := sm.users.FetchById(userId)
	if !ok { // ignore sessions for users that no longer exist
		return "", ok
	}
	user.IsAuthenticated = true

	id := uuid.New().String()
	sm.sessions.add(&Session{
		Id:        id,
		ExpiresAt: time.Now().Add(2 * 7 * 24 * time.Hour), // 2 weeks,
		User:      user,
	})

	return id, ok
}

func (sm *Manager) GetSession(sessionId string) (Session, bool) {
	sess, ok := sm.sessions.get(sessionId)
	if !ok {
		return Session{}, false
	}

	return Session{
		Id:        sess.Id,
		ExpiresAt: sess.ExpiresAt,
		User:      sess.User,
	}, true
}

func (sm *Manager) DeleteSession(id string) {
	sm.sessions.del(id)
}

// AddCookie creates a new session cookie.
// If there is no session with the given id, then the cookie will be deleted.
func (sm *Manager) AddCookie(w http.ResponseWriter, id string) bool {
	log.Printf("addCookie: session id %q\n", id)
	sess, ok := sm.sessions.get(id)
	if !ok {
		sm.DeleteCookie(w)
		return false
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sm.cookieName,
		Path:     "/",
		Value:    sess.Id,
		Expires:  sess.ExpiresAt,
		HttpOnly: true, // Set the HttpOnly flag to prevent client-side script access
		Secure:   true, // Set the Secure flag to ensure the cookie is only sent over HTTPS
		SameSite: http.SameSiteLaxMode,
	})

	return true
}

// GetCookie returns the session id from the request cookie.
func (sm *Manager) GetCookie(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(sm.cookieName)
	if err != nil {
		return "", false
	}
	return cookie.Value, true
}

// DeleteCookie deletes any session cookies.
func (sm *Manager) DeleteCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:    sm.cookieName,
		Path:    "/",
		MaxAge:  -1,
		Expires: time.Unix(0, 0),
	})
}

//// sessionFromCookie returns a session from the request cookie.
//func (fa *AuthorizationFactory) sessionFromCookie(cookie *http.Cookie) session {
//	if cookie == nil {
//		return session{
//			Roles: newRoles(),
//		}
//	}
//	sess, ok := fa.sessions.sessions[cookie.Value]
//	if !ok {
//		return session{
//			Roles: newRoles(),
//		}
//	}
//	return session{
//		Id:    sess.Id,
//		Name:  sess.Name,
//		Roles: sess.Roles.clone(),
//	}
//}
//
//func (fa *AuthorizationFactory) sessionFromId(id string) session {
//	sess, ok := fa.sessions.sessions[id]
//	if !ok {
//		return session{
//			Roles: newRoles(),
//		}
//	}
//	return session{
//		Id:    sess.Id,
//		Name:  sess.Name,
//		Roles: sess.Roles.clone(),
//	}
//}
//
//// sessionFromRequest returns a session from the request cookie.
//func (fa *AuthorizationFactory) sessionFromRequest(r *http.Request) session {
//	cookie, err := r.Cookie(fa.sessions.cookieName)
//	if err != nil {
//		return session{
//			Roles: newRoles(),
//		}
//	}
//	return fa.sessionFromCookie(cookie)
//}
