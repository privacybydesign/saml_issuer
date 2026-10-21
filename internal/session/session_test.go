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

func TestSessionExpiresAfterMaxAge(t *testing.T) {
	store, clock := newStore()
	id, err := store.Create()
	require.NoError(t, err)

	clock.Advance(session.MaxAge - time.Second)
	_, ok := store.Get(id)
	assert.True(t, ok)

	clock.Advance(time.Second)
	_, ok = store.Get(id)
	assert.False(t, ok)
	assert.False(t, store.Update(id, func(*session.Data) {}))
}

func TestNoSlidingExpiry(t *testing.T) {
	store, clock := newStore()
	id, err := store.Create()
	require.NoError(t, err)

	for range 7 {
		clock.Advance(time.Hour)
		require.True(t, store.Update(id, func(d *session.Data) { d.Pending[providerID] = session.Pending{Token: "x"} }))
	}
	clock.Advance(time.Hour)

	_, ok := store.Get(id)
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
	store.Update(id, func(d *session.Data) { d.Pending[providerID] = session.Pending{Token: "a"} })

	data, _ := store.Get(id)
	data.Pending[providerID] = session.Pending{Token: "changed"}
	data.Login[providerID] = []credential.Attribute{{Key: "id", Value: "x"}}

	again, _ := store.Get(id)
	assert.Equal(t, "a", again.Pending[providerID].Token)
	assert.Empty(t, again.Login)
}

func TestRotateMovesTheSession(t *testing.T) {
	store, clock := newStore()
	id, _ := store.Create()
	store.Update(id, func(d *session.Data) {
		d.Login[providerID] = []credential.Attribute{{Key: "id", Value: "jdoe"}}
	})
	clock.Advance(time.Hour)

	rotated, ok := store.Rotate(id)

	require.True(t, ok)
	assert.NotEqual(t, id, rotated)
	_, ok = store.Get(id)
	assert.False(t, ok, "the old ID must stop working")
	data, ok := store.Get(rotated)
	require.True(t, ok)
	assert.Equal(t, "jdoe", data.Login[providerID][0].Value)

	clock.Advance(session.MaxAge - time.Hour)
	_, ok = store.Get(rotated)
	assert.False(t, ok, "rotating must not extend the lifetime")
}

func TestDelete(t *testing.T) {
	store, _ := newStore()
	id, _ := store.Create()

	store.Delete(id)

	_, ok := store.Get(id)
	assert.False(t, ok)
}

func TestPurgeOnCreate(t *testing.T) {
	store, clock := newStore()
	old, _ := store.Create()
	clock.Advance(session.MaxAge)

	_, err := store.Create()

	require.NoError(t, err)
	_, ok := store.Get(old)
	assert.False(t, ok)
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
				if _, ok := d.Pending[providerID]; ok {
					delete(d.Pending, providerID)
					consumed.Add(1)
				}
			})
		})
	}
	wg.Wait()

	assert.EqualValues(t, 1, consumed.Load())
}
