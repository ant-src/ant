package value

import (
	"sort"
	"sync"
)

// Map is a hash map keyed by == that keeps insertion order (section 4). Reads hold its lock
// shared and mutations hold it exclusively. Entries carry a sequence number in insertion order,
// which a for loop uses as its cursor, so compaction never disturbs an iteration. Keys are hashed
// and compared with no lock held, since a key may contain the map itself and hashing or comparing
// it reads the map; ver counts the changes to the entries' positions, so a mutation can tell that
// the entry it found is still where it was.
type Map struct {
	mu      sync.RWMutex
	entries []mapEntry
	index   map[uint64][]int
	live    int
	seq     uint64
	ver     uint64
}

type mapEntry struct {
	key   Value
	value Value
	seq   uint64
	hash  uint64 // the key's hash when it was inserted
	dead  bool
}

// NewMap creates an empty map.
func NewMap() *Map {
	return &Map{index: map[uint64][]int{}}
}

// Len is the number of entries.
func (m *Map) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.live
}

// find returns the index of the entry whose key is == k, or -1, with k's hash and the version
// of the map it read. It must be called with no lock held; the caller locks the map and checks
// that ver is unchanged before it uses the index, and calls find again otherwise.
func (m *Map) find(k Value) (i int, h uint64, ver uint64) {
	h = Hash(k)
	m.mu.RLock()
	ver = m.ver
	var cands []int
	var keys []Value
	for _, i := range m.index[h] {
		if e := &m.entries[i]; !e.dead {
			cands = append(cands, i)
			keys = append(keys, e.key)
		}
	}
	m.mu.RUnlock()
	for j, key := range keys {
		if Equal(key, k) {
			return cands[j], h, ver
		}
	}
	return -1, h, ver
}

// Get returns the value stored under k.
func (m *Map) Get(k Value) (Value, bool) {
	for {
		i, _, ver := m.find(k)
		m.mu.RLock()
		if m.ver != ver {
			m.mu.RUnlock()
			continue
		}
		var v Value
		if i >= 0 {
			v = m.entries[i].value
		}
		m.mu.RUnlock()
		return v, i >= 0
	}
}

// Has reports whether k is a key.
func (m *Map) Has(k Value) bool {
	_, ok := m.Get(k)
	return ok
}

// Set inserts or replaces the value under k. A replaced entry keeps its position.
func (m *Map) Set(k, v Value) {
	m.insert(k, v, true)
}

// insert adds k with the value v and reports whether it was new; an existing entry gets v when
// replace is set and is left alone otherwise.
func (m *Map) insert(k, v Value, replace bool) bool {
	for {
		i, h, ver := m.find(k)
		m.mu.Lock()
		if m.ver != ver {
			m.mu.Unlock()
			continue
		}
		defer m.mu.Unlock()
		if i >= 0 {
			if replace {
				m.entries[i].value = v
			}
			return false
		}
		m.seq++
		m.ver++
		m.entries = append(m.entries, mapEntry{key: k, value: v, seq: m.seq, hash: h})
		m.index[h] = append(m.index[h], len(m.entries)-1)
		m.live++
		return true
	}
}

// Remove deletes the entry under k and returns its value.
func (m *Map) Remove(k Value) (Value, bool) {
	for {
		i, _, ver := m.find(k)
		m.mu.Lock()
		if m.ver != ver {
			m.mu.Unlock()
			continue
		}
		defer m.mu.Unlock()
		if i < 0 {
			return nil, false
		}
		v := m.entries[i].value
		m.entries[i].dead = true
		m.entries[i].value = nil
		m.live--
		m.ver++
		if m.live*2 < len(m.entries) && len(m.entries) > 8 {
			m.compact()
		}
		return v, true
	}
}

func (m *Map) compact() {
	entries := make([]mapEntry, 0, m.live)
	index := make(map[uint64][]int, m.live)
	for _, e := range m.entries {
		if e.dead {
			continue
		}
		index[e.hash] = append(index[e.hash], len(entries))
		entries = append(entries, e)
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
