#!/bin/sh
# Compares the Go interpreter (ant, no global lock) with the C compiler on the antc branch of
# ant-spec (antc, one global lock). Builds both, checks that every program prints the same output
# under each, then times them with hyperfine. ANTC_DIR is a checkout of the antc branch with
# `make runtime fixpoint` done; antc finds its runtime relative to the working directory, so
# builds run from there.
set -eu
HERE=$(cd "$(dirname "$0")" && pwd)
ANTC_DIR=${ANTC_DIR:-$HERE/../../../antc}
OUT=${OUT:-$(mktemp -d)}
mkdir -p "$OUT/bin" "$OUT/res"
(cd "$HERE/../.." && go build -o "$OUT/bin/ant" ./cmd/ant)
ANT="$OUT/bin/ant"
for f in "$HERE"/*.ant; do
  n=$(basename "$f" .ant)
  (cd "$ANTC_DIR" && bin/antc build "$f" -o "$OUT/bin/$n")
  [ "$("$ANT" run "$f" 3)" = "$("$OUT/bin/$n" 3)" ] || { echo "output differs: $n" >&2; exit 1; }
done
cd "$HERE"
for f in *.ant; do
  n=${f%.ant}
  case $n in par_*) continue ;; esac
  hyperfine -N -w 2 -r 15 --export-json "$OUT/res/$n.json" "$ANT run $f" "$OUT/bin/$n"
done
for n in par_fib par_map; do
  for k in 1 2 4 8 10 14; do
    hyperfine -N -w 1 -r 5 --export-json "$OUT/res/${n}_$k.json" "$ANT run $n.ant $k" "$OUT/bin/$n $k"
  done
done
echo "results in $OUT/res"
