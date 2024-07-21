// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package htmx

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"
)

type sessionManager_t struct {
	sync.RWMutex
	data    string
	pattern *regexp.Regexp
	// key is SHA-256 hash of session ID, value is session ID
	sessions map[string]session_t
	ttl      time.Duration
}

func newSessionManager(path string) *sessionManager_t {
	rxSessionId := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

	return &sessionManager_t{
		data:     path,
		pattern:  rxSessionId,
		sessions: make(map[string]session_t),
		ttl:      7 * 24 * time.Hour,
	}
}

func (sm *sessionManager_t) addSession(id string) {
	sm.Lock()
	defer sm.Unlock()
	hash := sha256.Sum256([]byte(id))
	hashStr := hex.EncodeToString(hash[:])
	sm.sessions[hashStr] = session_t{
		id:      id,
		expires: time.Now().Add(sm.ttl),
	}
	log.Printf("sm: add %q -> %q\n", id, hashStr)
}

func (sm *sessionManager_t) fromRequest(r *http.Request) session_t {
	cookie, err := r.Cookie("session")
	if err != nil {
		return session_t{}
	}
	return sm.getSession(cookie.Value)
}

func (sm *sessionManager_t) getSession(id string) session_t {
	hash := sha256.Sum256([]byte(id))
	hashStr := hex.EncodeToString(hash[:])
	sm.RLock()
	defer sm.RUnlock()
	return sm.sessions[hashStr]
}

func (sm *sessionManager_t) loadSessions() {
	sm.refresh()
}

func (sm *sessionManager_t) refresh() {
	sm.Lock()
	defer sm.Unlock()

	sm.sessions = make(map[string]session_t)

	// scan the data path for sessions and add them to the sessions map
	entries, err := os.ReadDir(sm.data)
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		matches := sm.pattern.FindStringSubmatch(entry.Name())
		// log.Printf("found session: %q: matches %d\n", entry.Name(), len(matches))
		if len(matches) != 1 {
			// not a session path
			continue
		}
		var clan struct {
			Id   string
			Clan string
		}
		if data, err := os.ReadFile(filepath.Join(sm.data, entry.Name(), "clan.json")); err != nil {
			log.Printf("warn: %q: %v\n", entry.Name(), err)
			continue
		} else if err = json.Unmarshal(data, &clan); err != nil {
			log.Printf("warn: %q: %v\n", entry.Name(), err)
			continue
		}
		hash := sha256.Sum256([]byte(entry.Name()))
		hashStr := hex.EncodeToString(hash[:])
		sm.sessions[hashStr] = session_t{
			clan:    clan.Clan,
			id:      entry.Name(),
			key:     hashStr,
			expires: time.Now().Add(sm.ttl),
		}
		log.Printf("session: load %q -> %q\n", entry.Name(), hashStr)
	}
}

func (sm *sessionManager_t) removeSession(id string) {
	sm.refresh()
}

type session_t struct {
	clan    string // clan name
	id      string // session ID
	key     string // hashed session ID
	expires time.Time
}

func (s session_t) isValid() bool {
	return time.Now().Before(s.expires)
}
