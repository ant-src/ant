#!/bin/sh
# coverage.sh runs the whole test suite with whole-program coverage and fails when statement
# coverage is below the thresholds: the first argument overall, the second per package.
# Usage: scripts/coverage.sh [overall-min] [package-min]
#
# Coverage is statement-weighted, computed from the profile: a package's figure is the share of
# its statements that some test ran, not an average over its functions. A block that several
# test binaries report is counted once.
set -eu
overall_min=${1:-90}
package_min=${2:-80}
profile=${COVERPROFILE:-$(mktemp)}
report=$(mktemp)
go test -coverpkg=./... -coverprofile="$profile" ./... > /dev/null
status=0
awk -v overall_min="$overall_min" -v package_min="$package_min" '
NR == 1 { next }
{
	block = $1
	n = $2
	split(block, loc, ":")
	pkg = loc[1]
	sub(/\/[^\/]*$/, "", pkg)
	sub(/^github.com\/ant-src\/ant\/?/, "", pkg)
	if (pkg == "") pkg = "(root)"
	if (!(block in seen)) { seen[block] = 1; total[pkg] += n; all += n }
	if ($3 > 0 && !(block in covered)) { covered[block] = 1; hit[pkg] += n; allhit += n }
}
END {
	status = 0
	for (pkg in total) {
		pct = 100 * hit[pkg] / total[pkg]
		mark = ""
		if (pct + 0.05 < package_min) { mark = "  BELOW " package_min "%"; status = 1 }
		printf "%-22s %5.1f%%%s\n", pkg, pct, mark
	}
	pct = 100 * allhit / all
	mark = ""
	if (pct + 0.05 < overall_min) { mark = "  BELOW " overall_min "%"; status = 1 }
	printf "%-22s %5.1f%%%s\n", "total", pct, mark
	exit status
}' "$profile" > "$report" || status=$?
sort "$report"
rm -f "$report"
exit $status
