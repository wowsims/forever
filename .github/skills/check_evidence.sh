#!/bin/sh
# Verifies the evidence and ruling ids (docs/evidence/README.md).
#
#   sh .github/skills/check_evidence.sh
#   make skills-check                    (runs it after the path checks)
#
# Checks four things:
#   1. no two files in docs/evidence share an EV- number
#   2. every EV- id a tracked file cites has its file in docs/evidence
#   3. every evidence file is cited outside itself, or carries a "Replaced by" line
#   4. no two rulings bullets share an R- id, and every R- id a tracked file cites has its bullet
#
# docs/evidence/README.md and Go tests are skipped: their ids are examples.

set -eu

repo=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
cd "$repo"

fail=0

# --- 1. unique numbers --------------------------------------------------
dupes=$(ls docs/evidence | grep -oE '^EV-[0-9]{4}' | sort | uniq -d)
for id in $dupes; do
	echo "evidence: $id names more than one file in docs/evidence"
	fail=1
done

# --- 2. every cited id has a file ---------------------------------------
cited=$(git grep -hoE 'EV-[0-9]{4}' -- . ':!docs/evidence/README.md' ':!*_test.go' | sort -u)
for id in $cited; do
	set -- docs/evidence/"$id"-*.md
	if [ ! -e "$1" ]; then
		echo "evidence: $id is cited but docs/evidence has no $id-*.md"
		fail=1
	fi
done

# --- 3. every file is cited ---------------------------------------------
for f in docs/evidence/EV-*.md; do
	[ -e "$f" ] || continue
	id=$(basename "$f" | grep -oE '^EV-[0-9]{4}')
	grep -q '^Replaced by: ' "$f" && continue
	if ! git grep -qE "$id\b" -- . ":!$f" ':!docs/evidence/README.md' ':!*_test.go'; then
		echo "evidence: nothing cites $f"
		fail=1
	fi
done

# --- 4. ruling ids ------------------------------------------------------
defined=$(grep -hoE '^- \*\*R-[0-9]{4}\*\*' .github/skills/*/references/rulings.md | grep -oE 'R-[0-9]{4}' | sort)
for id in $(echo "$defined" | uniq -d); do
	echo "rulings: $id heads more than one bullet"
	fail=1
done
for id in $(git grep -hoE '\bR-[0-9]{4}\b' -- . ':!docs/evidence/README.md' ':!*_test.go' | sort -u); do
	if ! echo "$defined" | grep -qx "$id"; then
		echo "rulings: $id is cited but no references/rulings.md bullet carries it"
		fail=1
	fi
done

exit $fail
