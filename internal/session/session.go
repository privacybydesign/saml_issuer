// Package session keeps login state on the server. The browser only holds an
// unguessable session ID.
package session

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"maps"
	"slices"
	"sync"
	"time"

	"github.com/privacybydesign/saml_issuer/internal/credential"
)

const (
	// MaxAge is how long a session lives, counted from its creation.
	MaxAge = 8 * time.Hour

	maxSessions = 50_000
	idBytes     = 32
)

var errFull = errors.New("too many sessions")

// Pending is a login that has been sent to an identity provider and not yet
// answered.
type Pending struct {
	// Token ties the answer to this session: the SAML request ID or the
	// OAuth state.
	Token string
	// ReturnTo is the page the user goes back to once logged in.
	ReturnTo string
}

// Data is the state of one session, by provider.
type Data struct {
	Pending map[string]Pending
	// Login holds the attributes of each provider the user logged in with.
	Login map[string][]credential.Attribute
}

type entry struct {
	expires time.Time
	data    Data
}

// Store holds sessions in memory.
type Store struct {
	mu       sync.Mutex
	now      func() time.Time
	sessions map[string]*entry
}

// NewStore returns an empty store that reads time from now.
func NewStore(now func() time.Time) *Store {
	return &Store{now: now, sessions: map[string]*entry{}}
}

// Create starts a session and returns its ID.
func (s *Store) Create() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.purge()
	if len(s.sessions) >= maxSessions {
		return "", errFull
	}

	id, err := newID()
	if err != nil {
		return "", err
	}
	s.sessions[id] = &entry{expires: s.now().Add(MaxAge), data: newData()}

	return id, nil
}

// Get returns a copy of the session, if it exists and has not expired.
func (s *Store) Get(id string) (Data, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	found := s.lookup(id)
	if found == nil {
		return Data{}, false
	}

	return found.data.clone(), true
}

// Update changes the session under the store's lock. It reports false when
// the session does not exist or has expired.
func (s *Store) Update(id string, change func(*Data)) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	found := s.lookup(id)
	if found == nil {
		return false
	}
	change(&found.data)

	return true
}

// Rotate gives the session a new ID, keeping its expiry, so that an ID that
// was known before login is useless after it.
func (s *Store) Rotate(id string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	found := s.lookup(id)
	if found == nil {
		return "", false
	}

	rotated, err := newID()
	if err != nil {
		return "", false
	}
	delete(s.sessions, id)
	s.sessions[rotated] = found

	return rotated, true
}

// lookup must be called with the lock held.
func (s *Store) lookup(id string) *entry {
	found, ok := s.sessions[id]
	if !ok {
		return nil
	}
	if !s.now().Before(found.expires) {
		delete(s.sessions, id)
		return nil
	}

	return found
}

// purge must be called with the lock held.
func (s *Store) purge() {
	now := s.now()
	for id, found := range s.sessions {
		if now.Before(found.expires) {
			continue
		}

		delete(s.sessions, id)
	}
}

func newID() (string, error) {
	raw := make([]byte, idBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func newData() Data {
	return Data{Pending: map[string]Pending{}, Login: map[string][]credential.Attribute{}}
}

func (d Data) clone() Data {
	copied := newData()
	maps.Copy(copied.Pending, d.Pending)
	for provider, attributes := range d.Login {
		copied.Login[provider] = slices.Clone(attributes)
	}

	return copied
}
