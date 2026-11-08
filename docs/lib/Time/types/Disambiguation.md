<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.Disambiguation

```ant signature
type Time.Disambiguation = Time.Compatible | Time.Earlier | Time.Later
```

Time.Disambiguation says which instant to take for a wall-clock time that a daylight-saving change
skipped or repeated in a zone. Such a time has no instant in a gap and two in an overlap;
[Time.possibleInstants](../possibleInstants.md) lists them.

```ant signature
type Time.Compatible = {}
```

Time.Compatible is the disambiguation that takes the later instant for a wall-clock time in a gap
and the earlier for one in an overlap. It is the default, and what the common date libraries do.

```ant signature
type Time.Earlier = {}
```

Time.Earlier is the disambiguation that takes the earlier instant for a wall-clock time in a gap or
an overlap.

```ant signature
type Time.Later = {}
```

Time.Later is the disambiguation that takes the later instant for a wall-clock time in a gap or an
overlap.

<!-- end generated -->
