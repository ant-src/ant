package value

import (
	"sort"
	"sync"
)

// Map is a hash map keyed by == that keeps insertion order (section 4). Reads hold its lock
// shared and mutations hold it exclusively. Entries carry a sequence number in insertion order,
// which a for loop uses as its cursor, so compaction never disturbs an iteration. The index maps
// a hash to the last entry with it, plus one; entries with the same hash are chained through
// next, so adding a key allocates nothing beyond the entry.
type Map struct {
	mu      sync.RWMutex
	entries []mapEntry
	index   map[uint64]int32
	live    int
	seq     uint64
	self    cell // holds the map, for slots that store it
}

type mapEntry struct {
	key   Value
	value Value
	seq   uint64
	next  int32 // the previous entry with the same hash, plus one; 0 ends the chain
	dead  bool
}

// NewMap creates an empty map.
func NewMap() *Map {
	m := &Map{index: map[uint64]int32{}}
	m.self.v = m
	return m
}

// Len is the number of entries.
func (m *Map) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.live
}

func (m *Map) find(k Value) (int, uint64) {
	h := Hash(k)
	for i := m.index[h]; i != 0; i = m.entries[i-1].next {
		e := &m.entries[i-1]
		if !e.dead && Equal(e.key, k) {
			return int(i - 1), h
		}
	}
	return -1, h
}

// Get returns the value stored under k.
func (m *Map) Get(k Value) (Value, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	i, _ := m.find(k)
	if i < 0 {
		return nil, false
	}
	return m.entries[i].value, true
}

// Has reports whether k is a key.
func (m *Map) Has(k Value) bool {
	_, ok := m.Get(k)
	return ok
}

// Set inserts or replaces the value under k. A replaced entry keeps its position.
func (m *Map) Set(k, v Value) {
	m.mu.Lock()
	defer m.mu.Unlock()
	i, h := m.find(k)
	if i >= 0 {
		m.entries[i].value = v
		return
	}
	m.seq++
	m.entries = append(m.entries, mapEntry{key: k, value: v, seq: m.seq, next: m.index[h]})
	m.index[h] = int32(len(m.entries))
	m.live++
}

// Remove deletes the entry under k and returns its value.
func (m *Map) Remove(k Value) (Value, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	i, _ := m.find(k)
	if i < 0 {
		return nil, false
	}
	v := m.entries[i].value
	m.entries[i].dead = true
	m.entries[i].value = nil
	m.live--
	if m.live*2 < len(m.entries) && len(m.entries) > 8 {
		m.compact()
	}
	return v, true
}

func (m *Map) compact() {
	entries := make([]mapEntry, 0, m.live)
	index := make(map[uint64]int32, m.live)
	for _, e := range m.entries {
		if e.dead {
			continue
		}
		h := Hash(e.key)
		e.next = index[h]
		entries = append(entries, e)
		index[h] = int32(len(entries))
	}
	m.entries, m.index = entries, index
}

// Next returns the first live entry whose sequence number is above after, for iteration by
// insertion order while the map may change. The cursor to pass next time is the returned seq.
func (m *Map) Next(after uint64) (key, value Value, seq uint64, ok bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	i := sort.Search(len(m.entries), func(i int) bool { return m.entries[i].seq > after })
	for ; i < len(m.entries); i++ {
		e := &m.entries[i]
		if !e.dead {
			return e.key, e.value, e.seq, true
		}
	}
	return nil, nil, 0, false
}

// Keys copies the keys out in insertion order.
func (m *Map) Keys() []Value {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Value, 0, m.live)
	for _, e := range m.entries {
		if !e.dead {
			out = append(out, e.key)
		}
	}
	return out
}

// Values copies the values out in insertion order.
func (m *Map) Values() []Value {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Value, 0, m.live)
	for _, e := range m.entries {
		if !e.dead {
			out = append(out, e.value)
		}
	}
	return out
}

// KeyValue is one entry of a map snapshot.
type KeyValue struct {
	Key, Value Value
}

// Entries copies the entries out in insertion order.
func (m *Map) Entries() []KeyValue {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]KeyValue, 0, m.live)
	for _, e := range m.entries {
		if !e.dead {
			out = append(out, KeyValue{e.key, e.value})
		}
	}
	return out
}
