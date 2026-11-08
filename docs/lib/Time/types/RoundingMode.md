<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.RoundingMode

```ant signature
type Time.RoundingMode = Time.Ceil | Time.Floor | Time.Expand | Time.Trunc | Time.HalfCeil | Time.HalfFloor | Time.HalfExpand | Time.HalfTrunc | Time.HalfEven
```

Time.RoundingMode says how [`Time.round`](../round.md) and
[`Time.roundDuration`](../roundDuration.md) round. The default is Time.HalfExpand.

```ant signature
type Time.Ceil = {}
```

Time.Ceil is the rounding mode toward positive infinity.

```ant signature
type Time.Floor = {}
```

Time.Floor is the rounding mode toward negative infinity.

```ant signature
type Time.Expand = {}
```

Time.Expand is the rounding mode away from zero.

```ant signature
type Time.Trunc = {}
```

Time.Trunc is the rounding mode toward zero.

```ant signature
type Time.HalfCeil = {}
```

Time.HalfCeil is the rounding mode to the nearest multiple, with a tie toward positive infinity.

```ant signature
type Time.HalfFloor = {}
```

Time.HalfFloor is the rounding mode to the nearest multiple, with a tie toward negative infinity.

```ant signature
type Time.HalfExpand = {}
```

Time.HalfExpand is the rounding mode to the nearest multiple, with a tie away from zero, as
Math.round does. It is the default.

```ant signature
type Time.HalfTrunc = {}
```

Time.HalfTrunc is the rounding mode to the nearest multiple, with a tie toward zero.

```ant signature
type Time.HalfEven = {}
```

Time.HalfEven is the rounding mode to the nearest multiple, with a tie to the even multiple.

<!-- end generated -->
