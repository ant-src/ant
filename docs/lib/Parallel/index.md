<!-- status: skeleton -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Parallel

Parallel runs work on goroutines and waits for all of it. Each call makes a scope context, a child
of ctx, and passes it to every callback; the call cancels it on every exit, success included, and
never cancels ctx. A call neither returns nor panics until every callback it started has ended,
deferred blocks included, and every action of a callback happens before the call returns, so a value
a callback wrote can be read after the call with no lock. If a panic leaves a callback, the call
cancels the scope, waits for the others, and raises that panic in the caller with its original
message and trace; a later panic in the same call is reported and dropped. A go inside a callback is
not owned by the call.

<!-- end generated -->
