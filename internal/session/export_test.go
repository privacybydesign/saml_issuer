package session

const MaxSessions = maxSessions

func (s *Store) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.sessions)
}
