<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.localZone

```ant signature
func Time.localZone(): Time.Zone
```

Time.localZone returns the host's zone. It is the zone the TZ environment variable names, else the
one the system's setting names; when neither names a zone of the database, it is the fixed offset
the host reports at the moment of the call.

<!-- end generated -->
