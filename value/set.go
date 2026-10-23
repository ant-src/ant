package value

// Set is a hash set keyed by == that keeps insertion order (section 4): a map whose values are
// all void, so it shares the map's lock, hashing, and iteration cursor.
type Set struct {
	m Map
}

// NewSet creates an empty set.
func NewSet() *Set {
	return &Set{m: Map{index: map[uint64][]int{}}}
}

// Len is the number of members.
func (s *Set) Len() int { return s.m.Len() }

// Has reports whether x is a member.
func (s *Set) Has(x Value) bool { return s.m.Has(x) }

// Add inserts x and reports whether it was new. A member already there keeps its position.
func (s *Set) Add(x Value) bool {
	s.m.mu.Lock()
	defer s.m.mu.Unlock()
	i, h := s.m.find(x)
	if i >= 0 {
		return false
	}
	s.m.seq++
	s.m.entries = append(s.m.entries, mapEntry{key: x, value: Void, seq: s.m.seq})
	s.m.index[h] = append(s.m.index[h], len(s.m.entries)-1)
	s.m.live++
	return true
}

// Remove deletes x and reports whether it was a member.
func (s *Set) Remove(x Value) bool {
	_, ok := s.m.Remove(x)
	return ok
}

// Next returns the first member whose sequence number is above after, as Map.Next does.
func (s *Set) Next(after uint64) (x Value, seq uint64, ok bool) {
	k, _, seq, ok := s.m.Next(after)
	return k, seq, ok
}

// Values copies the members out in insertion order.
func (s *Set) Values() []Value { return s.m.Keys() }
