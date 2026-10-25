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
	return s.m.insert(x, Void, false)
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

// Clone is a new set with the members of s, in its order.
func (s *Set) Clone() *Set {
	out := NewSet()
	out.AddAll(s)
	return out
}

// AddAll adds each member of t to s, in t's order.
func (s *Set) AddAll(t *Set) {
	for _, x := range t.Values() {
		s.Add(x)
	}
}
