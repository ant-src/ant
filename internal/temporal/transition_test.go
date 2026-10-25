package temporal

import "testing"

// scanTransitions finds every change of offset in [from, to) seconds by stepping an hour at a
// time and bisecting to the second, independently of Go's ZoneBounds, which NextTransition and
// PreviousTransition rely on.
func scanTransitions(z *Zone, from, to int64) []int64 {
	var out []int64
	off := func(sec int64) int { return z.offsetAt(Instant{sec: sec}) }
	const step = 3600
	prev := off(from)
	for lo := from; lo < to; lo += step {
		hi := lo + step
		cur := off(hi)
		if cur == prev {
			continue
		}
		a, b := lo, hi // off(a) == prev, off(b) != prev
		for b-a > 1 {
			m := a + (b-a)/2
			if off(m) == prev {
				a = m
			} else {
				b = m
			}
		}
		out = append(out, b)
		prev = cur
	}
	return out
}

// TestTransitionsAgainstScan pins NextTransition and PreviousTransition in zones with unusual
// histories, from 1960 into the years Go computes from the zone's last rule: every change of
// offset is found, in order, and none that is not one.
func TestTransitionsAgainstScan(t *testing.T) {
	zones := []string{
		"America/New_York", "Europe/London", "Europe/Dublin", "Europe/Vilnius", "Australia/Lord_Howe",
		"America/Santiago", "Pacific/Apia", "Africa/Casablanca", "America/Sao_Paulo", "Asia/Kolkata",
		"Antarctica/Troll", "Asia/Tehran", "Pacific/Chatham",
	}
	from := int64(-315619200) // 1960-01-01
	to := int64(2871763200)   // 2061-01-01
	for _, name := range zones {
		z, err := LoadZone(name)
		if err != nil {
			t.Fatal(err)
		}
		want := scanTransitions(z, from, to)
		var next []int64
		at := Instant{sec: from}.InZone(z)
		for {
			n, ok := at.NextTransition()
			if !ok || n.inst.sec >= to {
				break
			}
			if n.inst.nsec != 0 {
				t.Fatalf("%s: a transition at a fraction of a second: %s", name, n)
			}
			next = append(next, n.inst.sec)
			at = n
		}
		var prev []int64
		at = Instant{sec: to}.InZone(z)
		for {
			p, ok := at.PreviousTransition()
			if !ok || p.inst.sec <= from {
				break
			}
			prev = append([]int64{p.inst.sec}, prev...)
			at = p
		}
		for what, got := range map[string][]int64{"next": next, "previous": prev} {
			if len(got) != len(want) {
				t.Errorf("%s: %s finds %d transitions, the scan %d", name, what, len(got), len(want))
				continue
			}
			for i := range got {
				if got[i] != want[i] {
					t.Errorf("%s: %s transition %d at %s, the scan at %s", name, what, i,
						Instant{sec: got[i]}.InZone(z), Instant{sec: want[i]}.InZone(z))
					break
				}
			}
		}
	}
}
