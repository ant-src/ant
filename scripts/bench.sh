#!/bin/sh
# Compares the benchmarks of two commits: scripts/bench.sh OLD [NEW] [COUNT] [FILTER]
# NEW defaults to the working tree, COUNT to 10, FILTER (a -bench pattern) to every program.
# Each side runs in its own worktree, interleaved round by round so drift hits both alike, and
# benchstat prints the comparison. Install benchstat with
# go install golang.org/x/perf/cmd/benchstat@latest.
set -eu
old=$1
new=${2:-}
count=${3:-10}
filter=${4:-.}
root=$(git rev-parse --show-toplevel)
tmp=$(mktemp -d)
trap 'git -C "$root" worktree remove --force "$tmp/old" >/dev/null 2>&1 || true; rm -rf "$tmp"' EXIT
git -C "$root" worktree add --detach "$tmp/old" "$old" >/dev/null 2>&1
newdir=$root
if [ -n "$new" ]; then
  git -C "$root" worktree add --detach "$tmp/new" "$new" >/dev/null 2>&1
  newdir=$tmp/new
  trap 'git -C "$root" worktree remove --force "$tmp/old" >/dev/null 2>&1 || true; git -C "$root" worktree remove --force "$tmp/new" >/dev/null 2>&1 || true; rm -rf "$tmp"' EXIT
fi
(cd "$tmp/old" && go test -c -o "$tmp/old.test" ./bench)
(cd "$newdir" && go test -c -o "$tmp/new.test" ./bench)
i=0
while [ "$i" -lt "$count" ]; do
  (cd "$tmp/old/bench" && "$tmp/old.test" -test.run '^$' -test.bench "$filter" -test.benchmem -test.count 1) | grep '^Benchmark' >> "$tmp/old.txt"
  (cd "$newdir/bench" && "$tmp/new.test" -test.run '^$' -test.bench "$filter" -test.benchmem -test.count 1) | grep '^Benchmark' >> "$tmp/new.txt"
  i=$((i + 1))
done
benchstat "$tmp/old.txt" "$tmp/new.txt"
