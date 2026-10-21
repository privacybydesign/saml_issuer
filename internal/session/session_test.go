package session_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/privacybydesign/saml_issuer/internal/credential"
	"github.com/privacybydesign/saml_issuer/internal/session"
)

const providerID = "surfnet"

func newStore() (*session.Store, *clockwork.FakeClock) {
	clock := clockwork.NewFakeClockAt(time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC))

	return session.NewStore(clock.Now), clock
}

func TestSessionMaxAge(t *testing.T) {
	assert.Equal(t, 8*time.Hour, session.MaxAge)
}

func TestLoginExpiresAfterMaxAge(t *testing.T) {
	store, clock := newStore()
	id, _ := store.Create()
	rotated, ok := store.Rotate(id)
	require.True(t, ok)

	clock.Advance(session.MaxAge - time.Second)
	_, ok = store.Get(rotated)
	assert.True(t, ok)

	clock.Advance(time.Second)
	_, ok = store.Get(rotated)
	assert.False(t, ok)
	assert.False(t, store.Update(rotated, func(*session.Data) {}))
}

func TestPendingSessionIsShortLived(t *testing.T) {
	store, clock := newStore()
	id, err := store.Create()
	require.NoError(t, err)

	clock.Advance(5 * time.Minute)
	_, ok := store.Get(id)
	assert.True(t, ok, "the round trip to an identity provider takes minutes")

	clock.Advance(time.Hour)
	_, ok = store.Get(id)
	assert.False(t, ok, "a session without a login must not live for hours")
}

func TestNoSlidingExpiry(t *testing.T) {
	store, clock := newStore()
	id, _ := store.Create()
	rotated, _ := store.Rotate(id)

	for range 7 {
		clock.Advance(time.Hour)
		require.True(t, store.Update(rotated, func(d *session.Data) { d.Pending[providerID] = session.Pending{Token: "x"} }))
	}
	clock.Advance(time.Hour)

	_, ok := store.Get(rotated)
	assert.False(t, ok)
}

func TestSessionIDs(t *testing.T) {
	store, _ := newStore()
	seen := map[string]bool{}

	for range 100 {
		id, err := store.Create()
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(id), 43, "256 bits as base64")
		require.False(t, seen[id])
		seen[id] = true
	}
}

func TestUnknownSession(t *testing.T) {
	store, _ := newStore()

	_, ok := store.Get("nope")
	assert.False(t, ok)
	assert.False(t, store.Update("nope", func(*session.Data) {}))

	_, ok = store.Rotate("nope")
	assert.False(t, ok)
}

func TestGetReturnsACopy(t *testing.T) {
	store, _ := newStore()
	id, _ := store.Create()
	store.Update(id, func(d *session.Data) {
		d.Pending[providerID] = session.Pending{Token: "a"}
		d.Login[providerID] = []credential.Attribute{{Key: "id", Value: "jdoe"}}
	})

	data, _ := store.Get(id)
	data.Pending[providerID] = session.Pending{Token: "changed"}
	data.Login[providerID][0].Value = "changed"
	data.Login["other"] = []credential.Attribute{{Key: "id", Value: "x"}}

	again, _ := store.Get(id)
	assert.Equal(t, "a", again.Pending[providerID].Token)
	assert.Equal(t, "jdoe", again.Login[providerID][0].Value)
	assert.NotContains(t, again.Login, "other")
}

func TestRotateMovesTheSession(t *testing.T) {
	store, clock := newStore()
	id, _ := store.Create()
	store.Update(id, func(d *session.Data) {
		d.Login[providerID] = []credential.Attribute{{Key: "id", Value: "jdoe"}}
	})
	clock.Advance(time.Minute)

	rotated, ok := store.Rotate(id)

	require.True(t, ok)
	assert.NotEqual(t, id, rotated)
	_, ok = store.Get(id)
	assert.False(t, ok, "the old ID must stop working")
	data, ok := store.Get(rotated)
	require.True(t, ok)
	assert.Equal(t, "jdoe", data.Login[providerID][0].Value)
}

func TestRotateGrantsTheLoginLifetime(t *testing.T) {
	store, clock := newStore()
	id, _ := store.Create()
	clock.Advance(10 * time.Minute)

	rotated, ok := store.Rotate(id)
	require.True(t, ok)

	clock.Advance(session.MaxAge - time.Second)
	_, ok = store.Get(rotated)
	assert.True(t, ok, "the lifetime counts from the login")
}

func TestRotateAgainKeepsTheLifetime(t *testing.T) {
	store, clock := newStore()
	id, _ := store.Create()
	first, _ := store.Rotate(id)
	clock.Advance(time.Hour)

	second, ok := store.Rotate(first)
	require.True(t, ok)

	clock.Advance(session.MaxAge - time.Hour)
	_, ok = store.Get(second)
	assert.False(t, ok, "a second login must not extend the first")
}

func TestPurgeOnCreate(t *testing.T) {
	store, clock := newStore()
	old, _ := store.Create()
	loggedIn, _ := store.Create()
	loggedIn, _ = store.Rotate(loggedIn)
	clock.Advance(session.MaxAge)

	_, err := store.Create()

	require.NoError(t, err)
	assert.Equal(t, 1, store.Len(), "expired sessions must not stay in memory")
	_, ok := store.Get(old)
	assert.False(t, ok)
	_, ok = store.Get(loggedIn)
	assert.False(t, ok)
}

func fill(t *testing.T, store *session.Store, count int) []string {
	t.Helper()

	ids := make([]string, count)
	for i := range ids {
		id, err := store.Create()
		require.NoError(t, err)
		ids[i] = id
	}

	return ids
}

func TestFullStoreEvictsPendingSessions(t *testing.T) {
	store, clock := newStore()
	ids := fill(t, store, session.MaxSessions)
	loggedIn, _ := store.Rotate(ids[0])
	clock.Advance(time.Minute)

	fresh, err := store.Create()

	require.NoError(t, err, "visitors that only started a login must not lock others out")
	assert.Equal(t, session.MaxSessions, store.Len())
	_, ok := store.Get(fresh)
	assert.True(t, ok)
	_, ok = store.Get(ids[1])
	assert.False(t, ok, "the oldest pending session goes first")
	_, ok = store.Get(ids[2])
	assert.True(t, ok)
	_, ok = store.Get(loggedIn)
	assert.True(t, ok, "a login is never evicted for a pending session")
}

func TestFullStoreOfLogins(t *testing.T) {
	store, _ := newStore()
	for _, id := range fill(t, store, session.MaxSessions) {
		_, ok := store.Rotate(id)
		require.True(t, ok)
	}

	_, err := store.Create()

	assert.Error(t, err)
}

func TestUpdateIsAtomic(t *testing.T) {
	store, _ := newStore()
	id, _ := store.Create()
	store.Update(id, func(d *session.Data) { d.Pending[providerID] = session.Pending{Token: "once"} })

	var consumed atomic.Int32
	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() {
			store.Update(id, func(d *session.Data) {
				if _, ok := d.Pending[providerID]; !ok {
					return
				}

				delete(d.Pending, providerID)
				consumed.Add(1)
			})
		})
	}
	wg.Wait()

	assert.EqualValues(t, 1, consumed.Load())
}
