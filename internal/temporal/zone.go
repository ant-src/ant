package temporal

// The zone table comes from the IANA database's tzdata.zi and CLDR's windowsZones.xml:
//go:generate go run ./gen -tzdata /usr/share/zoneinfo/tzdata.zi -windows windowsZones.xml -o zones_gen.go

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Zone is a zone of the IANA database, or a fixed offset.
type Zone struct {
	id      string // as the database spells it, or the offset as ±HH:MM
	primary string // the zone a link names; the id itself otherwise
	loc     *time.Location
	fixed   bool
	offset  int // fixed zones: seconds east of UTC
}

var (
	zoneMu    sync.Mutex
	zoneCache = map[string]*Zone{}
)

// LoadZone resolves a name of the IANA database in any letter case, a link such as
// Asia/Calcutta included, or an offset ±HH, ±HH:MM or ±HHMM. Anything else is an error,
// the empty string and Go's "Local" included.
func LoadZone(id string) (*Zone, error) {
	if off, ok := parseOffsetZone(id); ok {
		return FixedZone(off), nil
	}
	spelled, ok := zoneNames[strings.ToLower(id)]
	if !ok {
		return nil, fmt.Errorf("unknown time zone %q", id)
	}
	zoneMu.Lock()
	defer zoneMu.Unlock()
	if z, ok := zoneCache[spelled]; ok {
		return z, nil
	}
	primary := spelled
	if target, ok := zoneLinks[spelled]; ok {
		primary = target
	}
	loc, err := time.LoadLocation(primary)
	if err != nil {
		return nil, fmt.Errorf("time zone %s: %v", spelled, err)
	}
	z := &Zone{id: spelled, primary: primary, loc: loc}
	zoneCache[spelled] = z
	return z, nil
}

// FixedZone is a zone at a fixed offset, in seconds east of UTC.
func FixedZone(offset int) *Zone {
	id := formatOffset(offset)
	return &Zone{id: id, primary: id, loc: time.FixedZone(id, offset), fixed: true, offset: offset}
}

// UTC is the zone UTC.
var UTC = mustZone("UTC")

func mustZone(id string) *Zone {
	z, err := LoadZone(id)
	if err != nil {
		panic(err)
	}
	return z
}

// parseOffsetZone reads ±HH, ±HH:MM or ±HHMM.
func parseOffsetZone(s string) (int, bool) {
	if len(s) < 3 || (s[0] != '+' && s[0] != '-') {
		return 0, false
	}
	rest := s[1:]
	h, m := rest, "0"
	switch {
	case len(rest) == 5 && rest[2] == ':':
		h, m = rest[:2], rest[3:]
	case len(rest) == 4:
		h, m = rest[:2], rest[2:]
	case len(rest) != 2:
		return 0, false
	}
	hh, err1 := strconv.Atoi(h)
	mm, err2 := strconv.Atoi(m)
	if err1 != nil || err2 != nil || len(h) != 2 || hh > 23 || mm > 59 || hh < 0 || mm < 0 {
		return 0, false
	}
	off := hh*3600 + mm*60
	if s[0] == '-' {
		off = -off
	}
	return off, true
}

// ID is the zone's identifier as given to LoadZone, in the database's spelling.
func (z *Zone) ID() string { return z.id }

// Fixed reports an offset zone.
func (z *Zone) Fixed() bool { return z.fixed }

// offsetAt is the zone's offset from UTC at an instant, in seconds.
func (z *Zone) offsetAt(i Instant) int {
	if z.fixed {
		return z.offset
	}
	_, off := time.Unix(i.sec, int64(i.nsec)).In(z.loc).Zone()
	return off
}

// TimeType is the name of the type's tag.
func (*Zone) TimeType() string { return "Time.Zone" }

// Key is the canonical form == compares: the zone a link names, so Asia/Calcutta equals
// Asia/Kolkata; an offset zone never equals a named one.
func (z *Zone) Key() string {
	if z.fixed {
		return "offset " + z.id
	}
	return "zone " + z.primary
}

// String is the identifier.
func (z *Zone) String() string { return z.id }

// LocalZone is the zone the host's configuration names: the TZ environment variable first,
// the system's setting second, and when neither names a zone of the database, the fixed
// offset the host reports at the moment of the call.
func LocalZone() *Zone {
	if tz := strings.TrimPrefix(os.Getenv("TZ"), ":"); tz != "" {
		if z, err := LoadZone(tz); err == nil {
			return z
		}
	}
	if name := systemZoneName(); name != "" {
		if z, err := LoadZone(name); err == nil {
			return z
		}
	}
	_, off := time.Now().Zone()
	return FixedZone(off)
}

// TzdataVersion is the version of the IANA database the zone table was generated from; the
// rules themselves are the host's or Go's.
func TzdataVersion() string { return tzdataVersion }
