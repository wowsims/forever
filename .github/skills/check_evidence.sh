#!/bin/sh
# Verifies the evidence and ruling ids (docs/evidence/README.md, docs/rulings/README.md).
#
#   sh .github/skills/check_evidence.sh
#   make skills-check                    (runs it after the path checks)
#
# For each kind, EV- in docs/evidence and R- in docs/rulings, checks four things:
#   1. no two files share a number
#   2. every id a tracked file cites has its file
#   3. every file is named <prefix>-NNNN-<slug>.md
#   4. every file is cited outside itself, or carries a "Replaced by" line
# and that every ruling states its evidence: EV- ids, pending, or not needed: <why>.
#
# Each folder's README.md and the hover test are skipped: their ids are examples.

set -eu

repo=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
cd "$repo"

fail=0

# check <kind> <prefix> <folder>
check() {
	kind=$1 prefix=$2 folder=$3
	# Every citation outside the examples, as <file>:<id>.
	cites=$(git grep -IowE "$prefix-[0-9]{4}" -- . ':!docs/evidence/README.md' ':!docs/rulings/README.md' ':!tools/spelldata/printer/lsp_hover_test.go' | sort -u)

	for id in $(ls "$folder" | grep -oE "^$prefix-[0-9]{4}" | sort | uniq -d); do
		echo "$kind: $id names more than one file in $folder"
		fail=1
	done

	for id in $(printf '%s\n' "$cites" | sed 's/.*://' | sort -u); do
		set -- "$folder/$id"-*.md
		if [ ! -e "$1" ]; then
			echo "$kind: $id is cited but $folder has no $id-*.md"
			fail=1
		fi
	done

	for f in "$folder/$prefix"-*.md; do
		[ -e "$f" ] || continue
		if ! id=$(basename "$f" | grep -oE "^$prefix-[0-9]{4}-"); then
			echo "$kind: $f is not named $prefix-NNNN-<slug>.md"
			fail=1
			continue
		fi
		id=${id%-}
		grep -q '^Replaced by: ' "$f" && continue
		if ! printf '%s\n' "$cites" | grep -F ":$id" | grep -qvF "$f:"; then
			echo "$kind: nothing cites $f"
			fail=1
		fi
	done
}

check evidence EV docs/evidence
check rulings R docs/rulings

for f in docs/rulings/R-*.md; do
	[ -e "$f" ] || continue
	if [ "$(grep -cE '^- Evidence: (EV-[0-9]{4}(, EV-[0-9]{4})*|pending|not needed: .+)$' "$f")" != 1 ]; then
		echo "rulings: $f needs one Evidence line: EV- ids, pending, or not needed: <why>"
		fail=1
	fi
done

[ "$fail" = 0 ] && echo "evidence and rulings: OK"
exit "$fail"
