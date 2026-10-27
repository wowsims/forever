#!/bin/sh
# Verifies that a skill still describes a tree that exists.
#
#   sh .github/skills/check_paths.sh .github/skills/<skill>
#   make skills-check                                          (every skill)
#
# Checks four things:
#   1. every repo path the skill names in backticks still exists
#   2. every references/*.md is routed to from SKILL.md, and every file SKILL.md
#      routes to exists (a references/ name written with a path in front belongs
#      to another skill and is checked as a path by check 1)
#   3. the frontmatter `name:` still matches the directory name
#   4. every `wowsims-*` skill the markdown names in backticks is a directory under .github/skills
#
# A path is only checked when it is unambiguous: it contains a slash and its
# first segment exists in the repo root, or it is a known root file, and it
# contains no glob, placeholder or URL (* < … space ://). One `{a,b}` group is
# expanded, so `dir/{a,b}.ts` checks both files. A trailing `:<line>` is
# stripped, so a `file.ts:123` citation is checked as the file.
# A path git ignores is skipped: a clean checkout lacks it by design.

set -eu

here=$(CDPATH= cd -- "$1" && pwd)
repo=$(CDPATH= cd -- "$here/../../.." && pwd)
cd "$repo"

# SKILL.md plus references/*.md, when the directory exists.
set -- "$here"/references/*.md
[ -e "$1" ] || set --
docs="$here/SKILL.md $*"

fail=0

expand_braces() {
	case $1 in
	*'{'*'}'*)
		pre=${1%%\{*}
		rest=${1#*\{}
		body=${rest%%\}*}
		post=${rest#*\}}
		echo "$body" | tr ',' '\n' | while read -r alt; do echo "$pre$alt$post"; done
		;;
	*) echo "$1" ;;
	esac
}

# --- 1. paths -------------------------------------------------------------
# shellcheck disable=SC2086
paths=$(grep -oh '`[^`]*`' $docs |
	tr -d '`' |
	grep -Ev '[*< ]|…|://' |
	grep -E '/|^(package\.json|tsconfig\.json|makefile|test-locales\.mjs|vite\.[a-z.-]*mts|vitest\.config\.mts|\.oxlintrc\.json|\.oxfmtrc\.json)$' |
	grep -Ev '^\.{0,2}/' |
	while read -r p; do expand_braces "$p"; done |
	sed -E 's/:[0-9]+$//' |
	while read -r p; do
		case $p in
		*/*) [ -e "$repo/${p%%/*}" ] || continue ;;
		esac
		echo "$p"
	done |
	sort -u)

# Each path that is not a file goes to git as p and as p/, so a directory-only
# pattern matches a directory the checkout lacks; a tracked file is never sent
# as p/, where a pattern like ui/**/*_auto_gen.ts would match it.
ignored=$(printf '%s\n' "$paths" |
	while read -r p; do
		echo "$p"
		[ -f "$p" ] || echo "$p/"
	done |
	git check-ignore --stdin |
	sed 'p; s:/$::')
[ -z "$ignored" ] || paths=$(printf '%s\n' "$paths" | grep -vxF -- "$ignored" || true)
count=$(printf '%s\n' "$paths" | grep -c . || true)
for p in $paths; do
	[ -e "$p" ] || {
		echo "MISSING PATH  $p"
		fail=1
	}
done

# --- 2. routing -----------------------------------------------------------
routed=$(grep -oE '(^|[^/a-z-])references/[a-z0-9_-]+\.md' "$here/SKILL.md" | grep -o 'references/.*' | sort -u)
for f in "$@"; do
	name="references/${f##*/}"
	printf '%s\n' "$routed" | grep -qxF -- "$name" || {
		echo "UNROUTED      $name is not mentioned in SKILL.md"
		fail=1
	}
done
for name in $routed; do
	[ -f "$here/$name" ] || {
		echo "DANGLING      SKILL.md routes to $name, which does not exist"
		fail=1
	}
done

# --- 3. frontmatter -------------------------------------------------------
declared=$(sed -n '2s/^name: *//p' "$here/SKILL.md")
[ "$declared" = "$(basename "$here")" ] || {
	echo "BAD NAME      frontmatter name '$declared' != directory '$(basename "$here")'"
	fail=1
}

# --- 4. skill names -------------------------------------------------------
# shellcheck disable=SC2086
for token in $(grep -ohE '`wowsims-[a-z-]+`' $docs | tr -d '`' | sort -u); do
	[ -d "$repo/.github/skills/$token" ] || {
		echo "UNKNOWN SKILL $token"
		fail=1
	}
done

[ "$fail" = 0 ] && echo "$(basename "$here"): OK ($count paths checked)"
exit "$fail"
