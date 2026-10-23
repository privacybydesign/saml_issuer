// Package session keeps login state on the server. The browser only holds an
// unguessable session ID.
package session

import (
	"container/list"
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
	// MaxAge is how long a login lives, counted from the login.
	MaxAge = 8 * time.Hour

	// pendingMaxAge is how long a session without a login lives. Anyone can
	// start one, so it only has to outlast the round trip to an identity
	// provider.
	pendingMaxAge = 30 * time.Minute

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
	id      string
	expires time.Time
	data    Data

	// queue is the list the session is in, and element its place there.
	queue   *list.List
	element *list.Element
}

func (e *entry) join(queue *list.List) {
	e.queue = queue
	e.element = queue.PushBack(e)
}

func (e *entry) leave() {
	e.queue.Remove(e.element)
}

// Store holds sessions in memory.
//
// A session starts without a login and only gets the full lifetime from
// Rotate. Both kinds are kept in a list ordered by expiry, so expired
// sessions are dropped from the front without a scan of the whole store.
type Store struct {
	mu       sync.Mutex
	now      func() time.Time
	sessions map[string]*entry

	// pending holds the sessions without a login, oldest first.
	pending list.List
	// active holds the sessions with a login, oldest first.
	active list.List
}

// NewStore returns an empty store that reads time from now.
func NewStore(now func() time.Time) *Store {
	return &Store{now: now, sessions: map[string]*entry{}}
}

// Create starts a session without a login and returns its ID. When the store
// is full, the oldest session without a login makes room, so that visitors who
// never log in cannot keep others out.
func (s *Store) Create() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.purge()
	if len(s.sessions) >= maxSessions && !s.evictOldestPending() {
		return "", errFull
	}

	id, err := newID()
	if err != nil {
		return "", err
	}
	created := &entry{id: id, expires: s.now().Add(pendingMaxAge), data: newData()}
	s.sessions[id] = created
	created.join(&s.pending)

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

// Rotate gives the session a new ID, so an ID known before login is useless
// after it. The first rotation is the login: the session then lives for
// MaxAge from now. Later ones keep the expiry.
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
	found.id = rotated
	s.sessions[rotated] = found

	if found.queue == &s.pending {
		found.leave()
		found.expires = s.now().Add(MaxAge)
		found.join(&s.active)
	}

	return rotated, true
}

// lookup must be called with the lock held.
func (s *Store) lookup(id string) *entry {
	found, ok := s.sessions[id]
	if !ok {
		return nil
	}
	if !s.now().Before(found.expires) {
		s.remove(found)
		return nil
	}

	return found
}

// purge must be called with the lock held.
func (s *Store) purge() {
	now := s.now()
	for _, queue := range []*list.List{&s.pending, &s.active} {
		for front := queue.Front(); front != nil; front = queue.Front() {
			oldest := front.Value.(*entry)
			if now.Before(oldest.expires) {
				break
			}

			s.remove(oldest)
		}
	}
}

// evictOldestPending drops the oldest session without a login and reports
// whether there was one. It must be called with the lock held.
func (s *Store) evictOldestPending() bool {
	front := s.pending.Front()
	if front == nil {
		return false
	}

	s.remove(front.Value.(*entry))

	return true
}

// remove must be called with the lock held.
func (s *Store) remove(found *entry) {
	found.leave()
	delete(s.sessions, found.id)
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
