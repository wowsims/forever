#!/usr/bin/env python3
"""Turns the spell data change report a `make db` log carries into Markdown for a pull request body.

Reads the log on stdin and writes one section per spell data group that changed, with the added,
changed and removed spells as a nested checklist, then the shared rows, the groups kept on their
committed build and the store's notes. Groups with no change share one line. The output stays under
GitHub's pull request body limit; the build log keeps the full report.

    python3 tools/spelldata_report_markdown.py < db.log
"""

import re
import sys

MAX_CHARS = 60000

HEADER = re.compile(r'^spelldata: (\S+) (.+?) -> (\S+): (\d+) added, (\d+) changed, (\d+) removed$')
ENTRY = re.compile(r'^spelldata:   ([+~-]) (.+)$')
SHARED = re.compile(r'^spelldata: shared row (.+)$')
NOT_PORTED = re.compile(r'^(talent files )?kept, not ported: (.+)$')
NOTE = re.compile(r'^spelldata: (.+)$')

KINDS = {'+': 'Added', '~': 'Changed', '-': 'Removed'}

SPELL_URL = 'https://www.wowhead.com/forever/spell='


def linked(entry):
    spell_id, _, rest = entry.partition(' ')
    if not spell_id.isdigit():
        return entry
    return f'[`{spell_id}`]({SPELL_URL}{spell_id}) {rest}'.rstrip()


def parse(lines):
    groups, shared, not_ported, notes = [], [], [], []
    current = None
    for line in lines:
        line = line.rstrip('\n')
        if m := HEADER.match(line):
            current = {'name': m[1], 'from': m[2], 'to': m[3], 'entries': {k: [] for k in KINDS}}
            groups.append(current)
        elif (m := ENTRY.match(line)) and current is not None:
            current['entries'][m[1]].append(m[2])
        elif m := SHARED.match(line):
            shared.append(m[1])
        elif m := NOT_PORTED.match(line):
            not_ported.append(('talent files' if m[1] else 'spell data') + ': ' + m[2])
        elif m := NOTE.match(line):
            notes.append(m[1])
    return groups, shared, not_ported, notes


def render(groups, shared, not_ported, notes):
    out = ['## Spell data changes', '']
    if not groups and not shared and not not_ported:
        out += ['No spell data change report in this run.', '']
    unchanged = []
    for g in groups:
        counts = {kind: len(entries) for kind, entries in g['entries'].items()}
        if not any(counts.values()):
            unchanged.append(g['name'])
            continue
        out += [f"### {g['name']}", '',
                f"{g['from']} → {g['to']}: {counts['+']} added, {counts['~']} changed, {counts['-']} removed", '']
        for kind, label in KINDS.items():
            if g['entries'][kind]:
                out.append(f'- [ ] {label}')
                out += [f'  - [ ] {linked(entry)}' for entry in g['entries'][kind]]
        out.append('')
    if unchanged:
        out += [f"Unchanged: {', '.join(unchanged)}", '']
    if shared:
        out += ['### Shared rows', ''] + [f'- [ ] {linked(row)}' for row in shared] + ['']
    if not_ported:
        out += ['### Kept, not ported', ''] + [f'- {line}' for line in not_ported] + ['']
    if notes:
        out += ['### Notes', ''] + [f'- {note}' for note in notes] + ['']
    return out


def capped(lines):
    text, size = [], 0
    for line in lines:
        if size + len(line) + 1 > MAX_CHARS:
            text += ['', '_Cut short to fit the pull request body; the build log has the full report._']
            break
        text.append(line)
        size += len(line) + 1
    return '\n'.join(text)


if __name__ == '__main__':
    print(capped(render(*parse(sys.stdin))))
