#!/bin/sh
# Verifies the evidence and ruling ids (docs/evidence/README.md, docs/rulings/README.md).
#
#   sh .github/skills/check_evidence.sh
#   make skills-check                    (runs it after the path checks)
#
# For each kind, EV- in docs/evidence and R- in docs/rulings, checks three things:
#   1. no two files share a number
#   2. every id a tracked file cites has its file
#   3. every file is cited outside itself, or carries a "Replaced by" line
#
# Each folder's README.md and Go tests are skipped: their ids are examples.

set -eu

repo=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
cd "$repo"

fail=0

# check <kind> <prefix> <folder>
check() {
	kind=$1 prefix=$2 folder=$3
	pattern="\\b$prefix-[0-9]{4}\\b"

	for id in $(ls "$folder" | grep -oE "^$prefix-[0-9]{4}" | sort | uniq -d); do
		echo "$kind: $id names more than one file in $folder"
		fail=1
	done

	for id in $(git grep -hoE "$pattern" -- . ':!docs/evidence/README.md' ':!docs/rulings/README.md' ':!*_test.go' | sort -u); do
		set -- "$folder/$id"-*.md
		if [ ! -e "$1" ]; then
			echo "$kind: $id is cited but $folder has no $id-*.md"
			fail=1
		fi
	done

	for f in "$folder/$prefix"-*.md; do
		[ -e "$f" ] || continue
		id=$(basename "$f" | grep -oE "^$prefix-[0-9]{4}")
		grep -q '^Replaced by: ' "$f" && continue
		if ! git grep -qE "$id\\b" -- . ":!$f" ':!docs/evidence/README.md' ':!docs/rulings/README.md' ':!*_test.go'; then
			echo "$kind: nothing cites $f"
			fail=1
		fi
	done
}

check evidence EV docs/evidence
check rulings R docs/rulings

exit $fail
