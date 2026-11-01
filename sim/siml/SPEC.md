# SIML v1

SIML (SimLanguage) is the text form of a WoWSims Forever rotation, gear set, or full sim setup. This document
is normative. The Go package `sim/siml` implements it and must pass the conformance corpus in §15. Every other
host reads and writes SIML through it: the site through `lib.wasm` or the server it runs on, the build's `.siml`
imports through `wowsimcli siml serve`, and editors through `wowsimcli siml lsp`.

"Must", "must not" and "may" are requirements. A _proto_ is a message of `proto/*.proto`. A _JSON_ value is
protojson: camelCase field names, enum values by name, 64-bit integers as strings, default values left out.

---

## 1. Model and file kinds

A SIML file is a set of **sections** over one `IndividualSimSettings`. Its file name's double extension says
which sections it may hold and what a loader returns:

| File name          | Holds                                           | Loaders return                                                    |
| ------------------ | ----------------------------------------------- | ----------------------------------------------------------------- |
| `x.siml`           | any sections: a full setup or any part of one   | `IndividualSimSettings`, plus the categories present              |
| `x.apl.siml`       | the rotation sections (§6), except `cooldowns:` | `APLRotation`                                                     |
| `x.gear.siml`      | the slot rows of a `gear:` section              | `EquipmentSpec` (17 items)                                        |
| `x.<section>.siml` | the body of `<section>`                         | `IndividualSimSettings` with only that section, plus its category |

`<section>` is any name in the section table (§5.1). In a section file the section's header is implied. The
header may still be written, but the printer leaves it out. Text with no file name, such as a paste, is read as
`x.siml`. The first written header of the implied section joins it, and another is `duplicate-section`. A
double extension that names no section is `unknown-section` at 1:1, and the file's body is not read.

`x.group.siml` is the exception: a group's header names the group, so each group is written under a header of its
own, which never joins the implied section, and the printer writes it. Every line before the first group header
but the `siml`, `spec` and `use` lines, a row, `var`, `def` or `when` line alike, belongs to no group and is
`misplaced-statement`, and that alone.

Every file may also hold the top-level lines of §3.3. A file holds names only, and the name table (§7) says what
they mean.

---

## 2. Lexical structure

**Source.** Source text is UTF-8. One U+FEFF, a byte order mark, at the start of the source is skipped: positions
count from the character after it, and `format` never writes it. Bytes that are not UTF-8 read as U+FFFD, one for
each maximal subpart of an ill-formed sequence, as the WHATWG decoder of browsers and Node reads them. `\r\n` is
read as `\n`. A tab counts as one column of indentation, and the printer writes spaces only.

**Columns.** In diagnostics, columns count UTF-16 code units and start at 1. Lines also start at 1.

**Tokens.**

| Token       | Pattern                                                                                                                                                | Notes                                                                              |
| ----------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------- |
| comment     | `//` to end of line, not followed by `/`                                                                                                               | file trivia: kept by `format`, never in the proto                                  |
| note        | `///` at the start of a line's text, then the rest of the line                                                                                         | see §10                                                                            |
| ident       | `[A-Za-z_][A-Za-z0-9_]*`                                                                                                                               | ASCII only                                                                         |
| int         | `[0-9]+`                                                                                                                                               | as an entry value                                                                  |
| literal     | `[0-9]+(\.[0-9]+)?` followed by one of `%`, `ms`, `us`, `µs`, `ns`, `s`, `m`, `h`, or nothing; or a Go compound duration such as `1m30s` or `1h2m3.5s` | no letter, digit or `_` may follow; the text is kept exactly (`5.00` stays `5.00`) |
| string      | a JSON string on one line                                                                                                                              | standard JSON escapes                                                              |
| raw         | `` ` `` JSON text `` ` ``                                                                                                                              | may span lines; a backtick inside is written as the JSON escape `\u0060`           |
| var         | `$` ident, or `${` name `}`: the name is any characters except `}`, `\` and newline, and the escapes `\}`, `\\` and `\n` for `}`, `\` and a line feed  |                                                                                    |
| placeholder | `%` ident, or `%{` name `}`, the name as in a var                                                                                                      |                                                                                    |
| punctuation | `>= <= == != < > + - * / ( ) [ ] , = . : #`                                                                                                            |                                                                                    |
| rejected    | `!`, `&&`, `&`, `\|\|`, `\|`                                                                                                                           | each gives a targeted error (§13)                                                  |

A character that starts no token is the error `bad-character`.

**Malformed tokens.** Each of these is one token with its own diagnostic:

- a run of digits, letters, `_`, `.` and `µ` that starts with a digit and reads as no literal (`5min`, `1e5`)
  is `bad-literal`;
- a string holding an escape JSON does not have, an escape of a lone surrogate (`"\ud800"`), which names no
  character, or a control character such as a tab, is `bad-character` at that escape or character;
- `$` or `%` followed by neither a name nor `{`, and `${` or `%{` with no `}` on its line, are `expected-name`;
- a `\` in `${…}` or `%{…}` followed by anything but `}`, `\` or `n` is `bad-character` at that escape;
- a string with no closing `"` on its line is `unclosed-string`, and raw JSON with no closing backtick is
  `unclosed-raw` at its backtick.

A `///` that follows other text on its line is a note too, and so `misplaced-note` (§10).

---

## 3. Statements and layout

### 3.1 Lines into statements

A file is a sequence of **statements**, each starting on a new line.

**When a statement continues.** It continues onto the next line in either of these cases:

- a `(`, `[` or backtick opened in it is still open;
- it is not a header (§3.2), and the next line that is not blank and not comment-only is indented deeper than
  the statement's first line and does not start with a header's tokens.

A `when` line (§6.9) ends at the first `:` that ends one of its lines with no bracket open.

**What ends a statement.** A blank line ends a statement unless a bracket is open. Comment-only lines never
decide continuation, and they are trivia.

**Indentation.** Indentation carries no other meaning, except that it says where a `when` body ends (§6.9). Every
statement after a header belongs to that section, whatever its indentation, until the next header.

### 3.2 Headers

A header is a statement that starts with one of these, and its first token must be at column 1:

- `ident :` — a section (§5.1);
- `group string :` — a rotation group (§6.1).

**A body line after the colon.** Text after the colon on the same line is one body line of that section: an
entry, a slot row, a row or a `var` line. That is how `buffs: greater_blessing_of_kings` works.

**Errors.**

- A header that is indented is `indented-header`.
- A header name that is not in the section table is `unknown-section`.
- A section the file kind does not hold is `section-not-allowed`.
- A section written twice is `duplicate-section`. Groups are told apart by their names; see `duplicate-group`.
- A `group:` header with no name is `parse`.
- The statements under a header that names no section, or under a `group:` with no name, are not read.

**A header never continues.** Its body lines may be indented any amount.

### 3.3 The first statements

| Statement                               | Means                                                                                                         | Rules                                                                                                                                                          |
| --------------------------------------- | ------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `siml 1`                                | the format's major version                                                                                    | Must be the first statement; comments and blank lines may come before it. Its absence is `missing-version`, and a major other than 1 is `unsupported-version`. |
| `spec <oneof>`                          | `Player.spec`: the oneof field of `Player`'s `spec`, such as `hunter`, `dps_warrior` or `retribution_paladin` | Required in `x.apl.siml` and `x.siml` (`spec-required`); see below for other kinds. It also sets `Player.class` (§8.4). An unknown name is `unknown-spec`.     |
| `talents "<string>"`                    | `Player.talents_string` (category t)                                                                          | In `x.siml` only.                                                                                                                                              |
| `use "<path>" [with <var> = <expr>, …]` | composition (§9)                                                                                              |                                                                                                                                                                |
| `var …`                                 | rotation value variables (§6.1)                                                                               |                                                                                                                                                                |
| `def …`                                 | rotation value macros (§6.10)                                                                                 |                                                                                                                                                                |

`siml`, `spec`, `talents` and `use` must come before the first header (`misplaced-statement`), and each may
appear once, except `use`.

**Where `spec` and `talents` go.** A file holds what its printer writes (§12.1). `spec` goes in `x.siml`, and in a
section file whose section reads the spec (`settings`, `simple`) or holds a rotation (category r: `apl`, `simple`,
`cooldowns`, `prepull`, `priority`, a group), whose names the spec's class reads (§7.1). `talents` goes in `x.siml`
only. Either line in any other file kind is `misplaced-statement`, so a section file holds only its own category
(§1).

A line whose first word is `siml`, `spec`, `talents`, `use`, `var` or `def` and whose next token is `=` is none of
these statements. It reads as a line of its section, so `var = 1` under `apl:` is an entry (`unknown-field`).

---

## 4. Grammar

The grammar is context-free given the current section, which the last header sets. A second pass, _typing_,
gives each node its meaning from the schema (§8).

```ebnf
file          = { trivia } , version , { statement | trivia } ;
trivia        = comment | blank ;
version       = "siml" , int ;

statement     = header | top | var_decl | def_decl | when_line | row | slot_row | entry ;

(* §3 *)
header        = ( ident , ":" | "group" , string , ":" ) , [ entry ] ;          (* column 1 only *)
top           = "spec" , ident
              | "talents" , string
              | "use" , string , [ "with" , var_bind , { "," , var_bind } ] ;
var_bind      = var , "=" , expr ;

(* §6: rotation sections *)
var_decl      = "var" , var , [ "=" , ( expr | raw ) ] ;
def_decl      = "def" , ident , [ "(" , [ ident , { "," , ident } ] , ")" ] , "=" , expr ;   (* §6.10 *)
when_line     = [ "#" ] , "when" , expr , ":" ;                   (* its body: §6.9 *)
row           = { note_line } , [ "#" ] , [ do_at ] , action ;    (* do_at present exactly in prepull: *)
note_line     = note ;                                            (* directly above, no blank line between *)
do_at         = unary ;                                           (* -22s, (1s + $x) *)
action        = step , [ "if" , expr ] ;
step          = "none"
              | "cast" , ref , [ "on" , unit ]
              | "channel" , ref , [ "on" , unit ] , [ "until" , expr ] , [ "recast" ]
              | "wait" , [ "until" ] , expr
              | "move" , expr , [ "recheck" ]
              | "move" , "for" , expr
              | "group" , string , [ "with" , ph_bind , { "," , ph_bind } ]    (* see note 1 *)
              | "schedule" , string , step
              | "sequence" , string , "[" , [ action , { "," , action } ] , "]"
              | call
              | raw ;
ph_bind       = placeholder , "=" , expr ;

expr          = conj , { "or" , conj } ;
conj          = neg , { "and" , neg } ;
neg           = "not" , neg | cmp ;
cmp           = sum , [ cmp_op , sum ] ;                          (* comparisons do not chain *)
cmp_op        = "==" | "!=" | ">=" | ">" | "<=" | "<" ;
sum           = prod , { ( "+" | "-" ) , prod } ;
prod          = unary , { ( "*" | "/" ) , unary } ;
unary         = [ "-" ] , literal | operand ;                     (* "-" only directly before a literal *)
operand       = "(" , expr , ")"
              | literal | string | "true" | "false" | "none"
              | var | placeholder | raw
              | [ prefix , "." ] , head , "(" , ( ref | string ) , ")" , "." , ident , [ args ]   (* accessor *)
              | prefix , "." , ident                              (* unit-prefixed kind *)
              | call ;
prefix        = ident | "none" ;                                  (* a unit word of §6.6; raw JSON is no prefix *)
head          = ident ;                                           (* one of the heads in §8.3 *)
call          = ident , [ args ] ;                                (* also "and" | "or" | "not" directly followed by "(" *)
args          = "(" , [ arg , { "," , arg } ] , ")" ;
arg           = [ ident , "=" ] , term ;
term          = action | expr | ref | unit | list ;               (* typed by the field *)
list          = "[" , [ term , { "," , term } ] , "]" ;

ref           = ident | ( "other" | "item" ) , "." , ident | "none" | raw ;   (* §7.2 *)
unit          = ident | "none" | raw ;                            (* the unit words of §6.6 *)

(* §5: setup sections *)
slot_row      = ident , ( ident , [ ident ] , [ "+" , ident ] | raw ) ;    (* gear: and item_swap:, first ident a slot word *)
entry         = flag_list | field_entry | message_entry | raw ;
flag_list     = flag , { "," , flag } ;                           (* every item a field name *)
flag          = ident , [ "=" , value ] ;                         (* mana_spring_totem=improved *)
field_entry   = ident , value , { "," , value } ;                 (* one field; several values only when repeated *)
message_entry = ident , args ;                                    (* a message value: targets(id=1, ...) *)
value         = literal | int | [ "-" ] , int | string | "true" | "false" | ident | unit | raw
              | ident , args ;                                    (* a nested message or indexed array *)
```

**Notes on the grammar.**

1. A `with` list continues past a comma only while the next token is a placeholder followed by `=`. That
   makes `sequence "S" [group "A" with %p = 1, cast x]` read as two actions.
2. `wait`, `move`, `cast`, `channel`, `group`, `schedule`, `sequence` and `none` are keywords in step position.
   `wait (1s + 2s) * 2` is `wait` of the expression `(1s + 2s) * 2`.
3. A row's tokens after a complete action that are not `if` give `expected-if`.
4. In operand position `sequence(` begins an accessor and `group` never does. In step position `sequence "` is
   the action.
5. The terms of a `field_entry` are values of one field. A line whose first item has no value, or is followed
   by `=`, is a `flag_list`; a flag with a value sets a flag-value enum to another of its values (§5.2), and
   is typed by its field as any entry value is. An entry's values, a flag's value and every argument are read
   as `term`s, and typing gives `wrong-type` for a term that does not fit its field. `siml`, `spec`, `talents`,
   `use`, `var` and `def` followed by `=` begin such a line, not a top-level line, a `var` line or a `def` line, so
   a field of one of those names is written `spec=arms` (`DpsWarrior.Rotation.spec`).
6. **Reserved words.** The name table never holds these as names (§7.4):
    - `siml spec talents use var with if on until recast recheck for and or not true false none other item when def`;
    - `cast channel wait move group sequence schedule`;
    - the heads (§8.3) and the unit words (§6.6).

    Section names are reserved only in header position.

7. **Terms.** A term that starts with `cast`, `channel`, `wait`, `move` or `group`, or with `schedule` or
   `sequence` and a string, is an action. A value followed by `if` is an action whose step is that value, which
   must be a call, raw JSON or `none`. In step position, `schedule` and `sequence` not followed by a string
   begin a call.
8. **Names.** In a ref, unit or operand position, `if`, `on`, `until`, `recast`, `recheck`, `for`, `with`, `and`, `or`,
   `not`, `cast`, `channel`, `wait`, `move`, `group` and `when` never read as a name. `and(`, `or(` and `not(` are calls
   only with no space before the `(`.
9. **Times.** In `prepull:`, a row that starts with an action keyword of note 7, or whose time would be followed
   by nothing or `if`, is `missing-time`, and the rest reads as its action. A row that is raw JSON alone is a
   whole row, its time inside it (§6.3), and is not `missing-time`. Elsewhere, a row that starts with a
   number, `-`, `(`, a string, a variable, a placeholder, `true` or `false` reads that value as a time and is
   `unexpected-time`.
10. **Depth.** A statement nests at most 256 deep, so that reading and printing it stay within a browser's stack.
    Each `(` and `[` goes one level deeper until it closes, and so do `not` and `schedule "…"` until what they hold
    ends; each `+`, `-`, `*` and `/` goes one level deeper for the rest of the statement, since an operator chain
    nests its left operand; and raw JSON goes as many levels deeper as its arrays and objects nest. A token that
    goes past 256 is `parse` ("the statement nests deeper than 256"), before raw JSON is read.

---

## 5. Setup sections

### 5.1 The section table

The table is data in `sim/siml/forms.json`, and the schema carries it (§8). Fields of a section are listed in
the order given here, and within each source in proto declaration order. The category letters are the site's
import categories (`SIM_CATEGORY_KEYS`).

| Section          | Fields, in order                                                                                                                                                                                                                                                                                                                                                                                    | Category | Body                  |
| ---------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | --------------------- |
| `settings`       | `Player`: `name`, `race`, `profession1`, `profession2`, `prof1_rank`, `prof2_rank`, `reaction_time_ms`, `channel_clip_delay_ms`, `in_front_of_target`, `distance_from_target`, `dark_intent_uptime`, `challenge_mode`, `weapon_skills`, `healing_model`; then the fields of the spec's `Options` other than `class_options`; then the fields of `class_options`; then `IndividualSimSettings.tanks` | m        | entries               |
| `gear`           | `Player.equipment` as slot rows; `Player.bonus_stats`; `Player.enable_item_swap`                                                                                                                                                                                                                                                                                                                    | g        | slot rows and entries |
| `item_swap`      | `Player.item_swap.items` as slot rows; `Player.item_swap.prepull_bonus_stats`                                                                                                                                                                                                                                                                                                                       | g        | slot rows and entries |
| `consumables`    | `Player.consumables`                                                                                                                                                                                                                                                                                                                                                                                | c        | entries               |
| `buffs`          | `Player.buffs`                                                                                                                                                                                                                                                                                                                                                                                      | x        | entries               |
| `party_buffs`    | `IndividualSimSettings.party_buffs`                                                                                                                                                                                                                                                                                                                                                                 | x        | entries               |
| `raid_buffs`     | `IndividualSimSettings.raid_buffs`; `IndividualSimSettings.target_dummies`                                                                                                                                                                                                                                                                                                                          | x        | entries               |
| `debuffs`        | `IndividualSimSettings.debuffs`                                                                                                                                                                                                                                                                                                                                                                     | x        | entries               |
| `encounter`      | `IndividualSimSettings.encounter`                                                                                                                                                                                                                                                                                                                                                                   | e        | entries               |
| `sim`            | `IndividualSimSettings.settings` (`SimSettings`); then `ep_weights_stats`, `ep_ratios`, `dps_ref_stat`, `heal_ref_stat`, `tank_ref_stat`, `reforge_settings`, `stat_caps`, `breakpoint_limits`                                                                                                                                                                                                      | u        | entries               |
| `apl`            | `Player.rotation.type`; `var` lines                                                                                                                                                                                                                                                                                                                                                                 | r        | entries and `var`     |
| `simple`         | the spec's `Rotation` message, stored as `Player.rotation.simple.spec_rotation_json`; `Player.rotation.simple.cooldowns` as the entry `cooldowns(...)`                                                                                                                                                                                                                                              | r        | entries               |
| `cooldowns`      | `Player.cooldowns`                                                                                                                                                                                                                                                                                                                                                                                  | r        | entries               |
| `prepull`        | `Player.rotation.prepull_actions`                                                                                                                                                                                                                                                                                                                                                                   | r        | rows                  |
| `priority`       | `Player.rotation.priority_list`                                                                                                                                                                                                                                                                                                                                                                     | r        | rows                  |
| `group "<name>"` | one `Player.rotation.groups` element                                                                                                                                                                                                                                                                                                                                                                | r        | `var` and rows        |

**Skipped fields.** These are never written or read: `Player.database`, every `api_version`, and every field of
type `UUID`. `Player.class` comes from `spec`
(§8.4).

**Categories.** A category is **present** when any of its sections is present, including empty, or when its
top-level line is present (`talents` for t). A present category's unwritten fields are zero in the file's
proto. The site's importer loads exactly the present categories, over the page's settings: a field the file leaves
out, or writes as zero, keeps the page's value, field by field in the setup, its player, its sim settings and its
cooldowns, while
any other message the file writes (the consumables, the encounter) replaces the page's whole.

**What a present section sets.** A present section sets its target message present, even when its body is
empty:

- `consumables`, `buffs`, `party_buffs`, `raid_buffs`, `debuffs`, `encounter`, `sim` (`settings`) and
  `cooldowns` set the message of that name;
- `item_swap` sets `Player.item_swap` with its 17 items;
- `gear` sets `Player.equipment` with its 17 items;
- `settings` sets the spec's `options` message, and its `class_options` message when the spec has one, as every
  spec's options in the site do;
- `simple` sets `rotation.simple` and its `spec_rotation_json`, which is `{}` when the section writes no field of
  the spec's `Rotation`.

**A section file's own section** is present when the file writes a statement of it, and otherwise when no file
the section file uses (§9) brings that section. So `x.apl.siml` holding only `use "y.apl.siml" with …` takes
`y`'s rotation without a `section-conflict`, and an empty `x.gear.siml` still yields its 17 empty items.

### 5.2 Entries

An entry line sets fields of the section's message. How each field type is written:

| Field type                                                                  | Written                                                                                        | Example                                                   | JSON                                            |
| --------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- | --------------------------------------------------------- | ----------------------------------------------- |
| `bool`                                                                      | the bare field name means true; `true` and `false` are also accepted                           | `cobra_reflexes`                                          | `"cobraReflexes": true`                         |
| enum                                                                        | the value's short name (§8.2), or the proto value name                                         | `race orc`                                                | `"race": "RaceOrc"`                             |
| an enum with a bare-flag value (`TristateEffect` → `TristateEffectRegular`) | bare means that value; otherwise its short name                                                | `mana_spring_totem`, `mana_spring_totem=improved`         | `"manaSpringTotem": "TristateEffectRegular"`    |
| `int32`, `uint32`, `int64`, `double`, `float`                               | a number with no unit                                                                          | `distance_from_target 11`                                 | `"distanceFromTarget": 11`                      |
| `string`                                                                    | a JSON string                                                                                  | `name "Pre-BiS - Arms"`                                   | `"name": "Pre-BiS - Arms"`                      |
| item, enchant or suffix ref field (§8.1)                                    | a name of that namespace (§7.2)                                                                | `offensive_pot major_frenzy_potion`                       | `"offensivePotId": 250943`                      |
| stat field (§8.1)                                                           | a `Stat` short name, or an integer that is no `Stat` value                                     | `stat_type2=attack_power`                                 | `"statType2": 14`                               |
| repeated scalar                                                             | values separated by commas                                                                     | `elixir elixir_of_the_mongoose, greater_mageblood_elixir` | `"elixirIds": [13452, 250341]`                  |
| message                                                                     | `field(arg, …)` with named arguments, bare bool flags, or both; `field()` is present but empty | `healing_model(hps=1200)`                                 | `"healingModel": {"hps": 1200}`                 |
| repeated message                                                            | one `field(...)` line per element                                                              | `targets(id=31146, …)`                                    | `"targets": [{…}]`                              |
| array indexed by an enum (§8.1)                                             | `field(<short name>=<number>, …)`                                                              | `stats(armor=3731, health=1299000)`                       | `"stats": [0, …, 3731, …]`, full length (§12.4) |
| `UnitReference`                                                             | a unit word (§6.6)                                                                             | `tanks player1`                                           | `"tanks": [{"type": "Player"}]`                 |
| `ActionID`                                                                  | a ref (§7)                                                                                     | `cooldowns(id=death_wish, timings=[30, 90])`              | `"id": {"spellId": 12328}`                      |
| map                                                                         | raw JSON for the whole map; a map has no other form                                            | ``min_consumable_lvl `{"6": 50}` ``                       | `"minConsumableLvl": {"6": 50}`                 |
| any                                                                         | raw JSON for the field's value                                                                 |                                                           | as written                                      |

- **Flag lists.** A flag list (`use_multi, use_arcane`) sets each named bool field to true, or each bare-flag
  enum to its flag value. A flag may carry a value (`mana_spring_totem=improved`, `pet_uptime=1`), which sets its
  field as an entry would.
- **Numbers.** A `float` field holds the double nearest the number, rounded to a float32, as the site holds it.
  The printer writes the shortest decimal that reads back to that float32, the nearer of two, and of two as near
  the one whose last digit is even (`0.123456789` reads as `0.12345679`). A number past the range of
  its field is `wrong-type`.
- **Stats.** A stat field reads a `Stat` short or proto name, or any integer; the printer writes the short name
  of a `Stat` value and the number otherwise.
- **Raw entries.** A line of raw JSON alone is a JSON object of fields of the section by JSON name, each set as
  its own entry would set it.
- **Messages in arguments.** A message field in a kind's or a message's arguments is written as in an entry,
  `field(...)` with the field's own name, or as raw JSON, or as `none` for the empty message. Each message of a
  list of messages is a `field(...)` argument of its own, as it is an entry line of its own in a section
  (`targets(…, target_inputs(…), target_inputs(…))`). The list may instead be written whole,
  `field=[field(…), …]`, but not both ways (`duplicate-field`). An array indexed by an enum is the argument
  `field(<short name>=<number>, …)`.
- **Field names.** A field is written by its proto name. A proto name that is not lower snake case is written
  as the snake case of its JSON name (`mhImbue_id` → `mh_imbue_id`). An item ref field whose name ends in
  `_id` or `_ids` drops that suffix when the shorter name names no other field of the message
  (`offensive_pot_id` → `offensive_pot`, `extra_ids` → `extra`). Both the short name and the full name are accepted, and the
  printer writes the short one.
- **Errors.** Unknown names are `unknown-field`. A field set twice is `duplicate-field`: a list of messages takes
  a line per message, and any other list one line of all its values. A value that does not fit is `wrong-type`.
  An unknown enum value is `unknown-enum`. A number with a unit is `bad-literal`.

### 5.3 Slot rows (`gear:`, `item_swap:`, `x.gear.siml`)

```
<slot> <item> [<suffix>] [+ <enchant>]
```

| Part          | Written                                                                                                                                           | JSON (`ItemSpec`)    |
| ------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------- |
| `<slot>`      | an `ItemSlot` short name: `head neck shoulder back chest wrist hands waist legs feet finger1 finger2 trinket1 trinket2 main_hand off_hand ranged` | the index in `items` |
| `<item>`      | an `item` name (§7.2), or raw `ItemSpec` JSON in place of the rest of the row                                                                     | `id`                 |
| `<suffix>`    | a `suffix` name                                                                                                                                   | `random_suffix`      |
| `+ <enchant>` | an `enchant` name (an effect id)                                                                                                                  | `enchant`            |

**Seventeen items.** A section with slot rows produces exactly 17 `ItemSpec`s, one per `ItemSlot`, and a slot
not written is `{}`. A slot written twice is `duplicate-slot`. An unknown slot word is `unknown-slot`.

**What a gear file may hold.** `x.gear.siml` holds slot rows only, so every line of its body reads as one. In
`x.siml`, `gear:` also takes the entries `bonus_stats(...)` and `enable_item_swap`. In a section with both, a
line whose first word is a slot word is a slot row, and any other line an entry.

### 5.4 The setup sections in short

| Section                                         | What it reads like                                                                                                                                                                    |
| ----------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `settings`                                      | `race orc`, `profession1 engineering`, `distance_from_target 11`, the spec options flattened (`quiver_bonus ancient_sinew_wrapped_lamina`, `cobra_reflexes`), `tanks player1`         |
| `consumables`                                   | `offensive_pot major_frenzy_potion`, `food swiftstrike_steak`, `extra goblin_sapper_charge, scroll_of_strength_iv`                                                                    |
| `buffs`, `party_buffs`, `raid_buffs`, `debuffs` | flag lists of the buff fields, `mana_spring_totem=improved` for a tristate, `gift_of_the_wild 1` for a rank select, `target_dummies 2` in `raid_buffs`                                |
| `encounter`                                     | `duration 180`, `duration_variation 5`, `execute_proportion_20 0.2`, one `targets(id=…, name="…", level=63, mob_type=mechanical, stats(…), …)` per target, written out field by field |
| `sim`                                           | `iterations 25000`, `ep_weights_stats(stats(agility=2.1))`, `dps_ref_stat attack_power`                                                                                               |
| `simple`                                        | the spec's `Rotation` fields (`time_to_weave 400`, `use_multi, use_arcane`, `playstyle weave`) and `cooldowns(hp_percent_for_defensives=0.4)`                                         |
| `cooldowns`                                     | `hp_percent_for_defensives 0.4`, `cooldowns(id=death_wish, timings=[30, 90])`                                                                                                         |

**Encoding `simple:`.** The `simple:` body is encoded into `spec_rotation_json` as **canonical rotation JSON**:

- the spec `Rotation` message's protojson;
- keys in proto declaration order;
- enum values by name, and default values left out;
- doubles in ECMAScript `Number#toString` form;
- no whitespace.

This is the text the site writes (`JSON.stringify(rotationToJson(...))` in `ui/app/preset_utils.ts`).

---

## 6. Rotation sections

### 6.1 Statements

| Statement                                     | Where                                    | JSON                                                                          |
| --------------------------------------------- | ---------------------------------------- | ----------------------------------------------------------------------------- |
| `var $name = <expr>` / `var ${Name} = <expr>` | top level, `apl:`, or a group            | `valueVariables[]` / `groups[g].variables[]`: `{"name": "Name", "value": V}`  |
| `var $name` / `var ${Name}`                   | top level, `apl:`, or a group            | a variable with no value: `{"name": "Name"}`                                  |
| `type <short name>`                           | `apl:`                                   | `"type"` (§6.2)                                                               |
| `prepull:`                                    | header                                   | `prepullActions`                                                              |
| `priority:`                                   | header                                   | `priorityList`                                                                |
| `group "Name":`                               | header                                   | a `groups[]` element `{"name": "Name", "actions": [...], "variables": [...]}` |
| a row                                         | under `prepull:`, `priority:` or a group | an `APLPrepullAction` or an `APLListItem`                                     |
| `when <expr>:` and its body                   | under `priority:` or a group             | its body's rows, each with the block's condition joined to its own (§6.9)     |
| `def name(p, …) = <expr>`                     | top level or `apl:`                      | none: each call reads as the def's value (§6.10)                              |

**Where `var` may appear.** A `var` line at top level or under `apl:` is a rotation value variable. Under a
group header it is one of the group's `variables`. Anywhere else, a `when` body among those places, it is
`misplaced-var`.

**Duplicates.** A duplicate group name is `duplicate-group`. A duplicate variable name in one scope is
`duplicate-variable`, and so is a variable outside the groups that this file and a used file both declare (§9).

### 6.2 Rotation type

| The file has                                                                     | `Player.rotation.type`          |
| -------------------------------------------------------------------------------- | ------------------------------- |
| an explicit `type` entry in `apl:`                                               | that value                      |
| else a `simple:` section                                                         | `TypeSimple`                    |
| else any rotation section or rotation `var` line, or the file is an `x.apl.siml` | `TypeAPL`                       |
| no rotation section                                                              | the rotation category is absent |

The rotation sections are the sections of category r other than `cooldowns:`, which writes `Player.cooldowns` and
leaves the rotation alone, so a file of cooldowns alone does not replace a rotation on import. In a file that uses
others (§9),
the type is read over all of them: an explicit `type` in any of them, else a `simple:` in any of them, else
`TypeAPL`.

**No `TypeUnknown`.** Inference never yields `TypeUnknown` for a present rotation category; only an explicit
`type unknown` does. An importer must check this, because `Player.fromProto` turns `TypeUnknown` into
`TypeAuto` (`ui/sim/player/player.ts:1511`).

**When the printer writes `type`.** Only when the inferred type differs from the proto's, so a rotation with
no type is written with `type unknown`.

### 6.3 Rows

```
[/// note lines]
[# ][<do_at> ]<step>[ if <expr>]
```

| Part                                                                                          | JSON                                            |
| --------------------------------------------------------------------------------------------- | ----------------------------------------------- |
| `# `                                                                                          | `"hide": true`                                  |
| `<do_at>`, required in `prepull:` and forbidden elsewhere (`missing-time`, `unexpected-time`) | `APLPrepullAction.do_at_value`                  |
| `<step>`                                                                                      | `action` without `condition`                    |
| `if <expr>`                                                                                   | `action.condition`                              |
| `///` lines directly above a `priority:` or group row                                         | `APLListItem.notes`, the lines joined with `\n` |

**`#` must hide a row.** A `#` not followed by a row is `bad-hidden-row`, with the message "`#` hides a row;
write `//` for a comment".

**Raw rows.** A row that is one raw token alone, with no `#`, time, condition or notes, is the JSON of the whole
row: the `APLPrepullAction` in `prepull:`, its time inside it, and the `APLListItem` elsewhere. Raw JSON in any
other step position is the action (§6.4). The printer writes a raw row for a row no other form writes, such as
one whose notes cannot be `///` lines (§12.3) or a prepull row with no time.

**A time outside `prepull:`** is `unexpected-time`, and typing does not read it.

### 6.4 Short forms of actions

| SIML                                  | JSON (`APLAction`, without `condition`)                                                                                                                                                                                                                              |
| ------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `none`                                | `{}`                                                                                                                                                                                                                                                                 |
| `cast x`                              | `{"castSpell": {"spellId": X}}`                                                                                                                                                                                                                                      |
| `cast x on u`                         | `{"castSpell": {"spellId": X, "target": U}}`                                                                                                                                                                                                                         |
| `channel x [on u] [until v] [recast]` | `{"channelSpell": {"spellId": X, "target": U, "interruptIf": V, "allowRecast": true}}`                                                                                                                                                                               |
| `wait v`                              | `{"wait": {"duration": V}}`                                                                                                                                                                                                                                          |
| `wait until v`                        | `{"waitUntil": {"condition": V}}`                                                                                                                                                                                                                                    |
| `move v`                              | `{"move": {"rangeFromTarget": V}}`                                                                                                                                                                                                                                   |
| `move v recheck`                      | `{"move": {"rangeFromTarget": V, "recheckWhileMoving": true}}`                                                                                                                                                                                                       |
| `move for v`                          | `{"moveDuration": {"duration": V}}`                                                                                                                                                                                                                                  |
| `group "N"`                           | `{"groupReference": {"groupName": "N"}}`                                                                                                                                                                                                                             |
| `group "N" with %p = v, %q = w`       | `{"groupReference": {"groupName": "N", "variables": [{"name": "p", "value": V}, {"name": "q", "value": W}]}}`                                                                                                                                                        |
| `schedule "0s, 60s" <step>`           | `{"schedule": {"schedule": "0s, 60s", "innerAction": STEP}}` (the inner step takes no `if`; the row's `if` is the schedule's)                                                                                                                                        |
| `sequence "N" [a, b]`                 | `{"sequence": {"name": "N", "actions": [A, B]}}` (each item an action, `if` allowed)                                                                                                                                                                                 |
| `<kind>(…)` or bare `<kind>`          | any action kind, by §8.3: `reset_sequence("HHA")`, `autocast_other_cooldowns`, `strict_sequence(item_swap(swap1), cast shield_wall)`, `cast_all_stat_buff_cooldowns(stat_type2=attack_power, stat_type3=haste_rating)`, `multidot(rend, max_dots=3, max_overlap=2s)` |
| raw                                   | the action's JSON; with `if`, it holds no `condition` (`duplicate-field`)                                                                                                                                                                                            |

`X` is a ref (§7), `U` a unit (§6.6), and `V` and `W` values.

### 6.5 Values

| SIML                                                 | JSON (`APLValue`)                                           |
| ---------------------------------------------------- | ----------------------------------------------------------- |
| `1.5s`, `400ms`, `5.00`, `30%`, `-22s`, `1m30s`, `7` | `{"const": {"val": "<the text exactly>"}}`                  |
| `true`, `false`                                      | `{"const": {"val": "true"}}`, `{"const": {"val": "false"}}` |
| `"text"`                                             | `{"const": {"val": "text"}}`                                |
| `none`                                               | `{}`                                                        |
| `$name`, `${Name}`                                   | `{"variableRef": {"name": "Name"}}` (§7.5)                  |
| `%name`, `%{Name}`                                   | `{"variablePlaceholder": {"name": "Name"}}`                 |
| `a and b and c`                                      | `{"and": {"vals": [A, B, C]}}`; one `and` with three values |
| `a or b`                                             | `{"or": {"vals": [A, B]}}`                                  |
| `and(x)`, `or()`                                     | an `and` or `or` with fewer than two values                 |
| `not a`                                              | `{"not": {"val": A}}`                                       |
| `a == b`, `!=`, `>=`, `>`, `<=`, `<`                 | `{"cmp": {"op": "OpEq"…"OpLt", "lhs": A, "rhs": B}}`        |
| `a + b`, `-`, `*`, `/`                               | `{"math": {"op": "OpAdd"…"OpDiv", "lhs": A, "rhs": B}}`     |
| `max(a, b)`, `min(a, b)`                             | `{"max": {"vals": [A, B]}}` (§8.3, list kinds)              |
| `[u.]head(x).prop[(f=v, …)]`                         | the accessor kind (§6.7)                                    |
| `u.kind`                                             | a kind whose only field is a unit (§6.8)                    |
| `kind`, `kind(…)`                                    | any value kind, by §8.3                                     |
| `name`, `name(a, …)` for a def                       | the def's value, its arguments in place (§6.10)             |
| raw                                                  | the value's JSON                                            |

**Precedence**, loosest first:

1. `or`
2. `and`
3. `not`
4. comparisons
5. `+` and `-`
6. `*` and `/`
7. operands

`and`, `or`, `+`, `-`, `*` and `/` are left-associative. `not a == b` means `not (a == b)`. Parentheses only
change the JSON when they keep an `and` inside an `and`, or an `or` inside an `or`, apart, or when they
override precedence.

**Number literals.** A bare number compared with a time counts as seconds in the sim. A percentage needs `%`.

### 6.6 Units

| SIML                             | `UnitReference` JSON                                                                     |
| -------------------------------- | ---------------------------------------------------------------------------------------- |
| `self`                           | `{"type": "Self"}`                                                                       |
| `target`                         | `{"type": "CurrentTarget"}`                                                              |
| `previous_target`, `next_target` | `{"type": "PreviousTarget"}`, `{"type": "NextTarget"}`                                   |
| `all_targets`, `all_players`     | `{"type": "AllTargets"}`, `{"type": "AllPlayers"}`                                       |
| `unknown_unit`                   | `{"type": "Unknown"}`, the default, which reads as `{}`; the printer writes `none`       |
| `targetN`, `playerN` (N ≥ 1)     | `{"type": "Target", "index": N-1}`, with `index` left out when N is 1; `Player` likewise |
| `petN`                           | `{"type": "Pet", "index": N-1, "owner": {"type": "Self"}}`                               |
| `none`                           | `{}`                                                                                     |
| raw                              | anything else, such as a pet with another owner                                          |

`UnitReference.index` is an `int32`, so N − 1 must fit one: `target2147483649` is `wrong-type`.

### 6.7 Accessors

An accessor is a value kind named `<head>_<prop>` (§8.3). It is written:

```
[<unit>.]<head>(<subject>).<prop>[(<field>=<value>, …)]
```

- The **subject** fills the kind's one `ActionID` field, or its one `string` field for the `sequence` head.
- The **unit prefix** fills its one `UnitReference` field.
- The other fields go on the property as named arguments, or bare for bools:
  `target.aura(sunder_armor).num_stacks(include_reaction_time=true)`.

Arguments written inside the head's parentheses are `args-on-head`.

The heads the v1 protos derive, each with a few of its properties. The list is not normative: §8.3 is, so a new
`aura_foo` kind is `aura(x).foo` with no edit. The `wowsims-siml` skill (`.github/skills/wowsims-siml/SKILL.md`)
lists every property with its arguments, written from the schema (`wowsimcli siml llm`, which `make siml` runs).

| Head         | Subject                    | Unit prefix fills                     | Properties, such as (the kind is `<head>_<prop>`)                                   |
| ------------ | -------------------------- | ------------------------------------- | ----------------------------------------------------------------------------------- |
| `aura`       | `aura_id`                  | `source_unit`                         | `is_active(include_reaction_time)`; `remaining_time`; `num_stacks`; `duration`      |
| `dot`        | `spell_id`                 | `target_unit`                         | `is_active`; `remaining_time`; `tick_frequency`; `percent_increase(use_base_value)` |
| `spell`      | `spell_id`                 | `target_unit`, where the kind has one | `is_ready`; `time_to_ready`; `cast_time`; `current_cost`; `expected_dpet`           |
| `boss_spell` | `spell_id`                 | `target_unit`                         | `time_to_ready`; `is_casting`                                                       |
| `sequence`   | `sequence_name` (a string) | (none)                                | `is_complete`; `is_ready`; `time_to_ready`                                          |

**Example.** In a hunter's file, `target.aura(hunters_mark).is_active` is
`{"auraIsActive": {"sourceUnit": {"type": "CurrentTarget"}, "auraId": {"spellId": 14325}}}`: the hunter's
`target_aura` namespace names spell 14325 `hunters_mark`.

### 6.8 Other derived value forms

These rules are in §8.3; the v1 instances are listed for reference.

| Form            | Rule                                                           | v1 instances                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| --------------- | -------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| bare word       | a kind with no fields                                          | `current_time`, `current_time_percent`, `remaining_time`, `remaining_time_percent`, `number_targets`, `current_rage`, `current_energy`, `current_combo_points`, `current_generic_resource`, `current_mana_regen`, `max_health`, `max_mana`, `max_combo_points`, `max_energy`, `max_rage`, `time_to_next_energy_tick`, `gcd_is_ready`, `gcd_time_to_ready`, `channel_clip_delay`, `input_delay`, `front_of_target`, `protection_paladin_damage_taken_last_global` |
| unit prefix     | a value kind whose only field is a `UnitReference`: `[u.]kind` | `current_health`, `current_health_percent`, `current_mana`, `current_mana_percent`, `unit_is_moving`, `unit_distance`, `boss_current_target`, `warlock_assigned_curse_is_active`                                                                                                                                                                                                                                                                                 |
| plain arguments | a kind whose only field is repeated                            | `max(a, b)`, `min(a, b)`, `and(x)`, `or(x)`, `strict_sequence(a, b)`                                                                                                                                                                                                                                                                                                                                                                                             |
| generic         | any other kind                                                 | `is_execute_phase(e20)`, `auto_time_to_next(main_hand_auto)`, `auto_time_since_last(main_hand_auto)`, `action_group_used("Stance Dance ON/OFF")`, `multiple_cd_usages(x, target_spell_id=y, offset=2s)`, `selected_potion(x)`, `energy_time_to_target(50)`, `totem_remaining_time(air)`                                                                                                                                                                          |

### 6.9 `when` blocks

```
priority:
  cast execute if is_execute_phase(e20)
  when not is_execute_phase(e20) and aura(battle_stance).is_active:
    cast rend if dot(rend).remaining_time <= 1.5s
    cast overpower
    # cast spearing_strike
  cast whirlwind
```

reads exactly as

```
priority:
  cast execute if is_execute_phase(e20)
  cast rend if not is_execute_phase(e20) and aura(battle_stance).is_active and dot(rend).remaining_time <= 1.5s
  cast overpower if not is_execute_phase(e20) and aura(battle_stance).is_active
  # cast spearing_strike if not is_execute_phase(e20) and aura(battle_stance).is_active
  cast whirlwind
```

A `when` line is followed by its **body**: the statements after it whose first line is indented deeper than the
`when` line. The body ends at the first statement indented no deeper than the `when` line, at a header, or at the end
of the file. It holds rows, other `when` blocks and comments. No block reaches the proto: each row of a body is a row
of the list, and `print` writes the rows flat.

**The line.** `when`, or `# when` (below), an expression and `:`. The statement ends at the first `:` that ends one of
its lines with no bracket open (§3.1), so a long condition may go on continuation lines; one that ends otherwise, with
no `:` or with more text after it, is `parse` ("a when line ends with :"). `when` followed by `=` begins no `when`
line.

**Where.** Under `priority:` and under a group header. A `when` line anywhere else is `misplaced-when`, at its first
token, and neither it nor its body is read. `when` reads as no value (§4 note 8), so a `when` inside a
`sequence "N" [...]` or a `with` list does not parse. A `var` line in a body is `misplaced-var`. A `when` line with
nothing in its body is the warning `empty-when`.

**What the rows read as.** Each row of a body reads with its condition joined to the block's:

- the _conjuncts_ of an expression are the values of an `and` written without parentheses at its top, and otherwise
  the expression alone;
- a row with condition `C` in blocks with conditions `W1`, the outermost, to `Wn` reads with the condition
  `{"and": {"vals": conj(W1) ++ … ++ conj(Wn) ++ conj(C)}}`;
- a row with no `if` reads the same way without `conj(C)`;
- a row whose join holds one condition, as a row with no `if` in one block does, reads with that condition as
  written: no `and` is built around it.

So `when a and b:` over `cast x if c and d` gives the JSON of `cast x if a and b and c and d`, `when a or b:` over
`cast x if c` that of `cast x if (a or b) and c`, and `when (a and b):` over `cast x if c` that of
`cast x if (a and b) and c`.

**Hidden blocks.** `# when …:` hides the block: every row in it, nested blocks' included, reads with `"hide": true`.
A `#` on a row inside works as it does anywhere else.

**Raw rows.** A raw row (§6.3) is the whole row's JSON in a body too: the block adds neither its condition nor
`"hide"` to it.

**Notes and positions.** `///` lines directly above a row of a body are that row's notes, and `///` lines above a
`when` line are `misplaced-note` (§10). Each row of a body keeps its own place in its list, in the order the text
writes the rows, so a row's index is the one the flat text gives it.

`format` keeps every block (§12.3).

### 6.10 Defs

```
var ${Move duration} = 1s * (($ranged_range - $melee_range) / 7)
def fits(x) = spell(x).cast_time + $move_duration <= $time_to_next_weave or not $melee_weave

priority:
  cast multi_shot if $use_multi_shot and fits(multi_shot)
  cast aimed_shot if $use_aimed_shot and fits(aimed_shot)
```

reads as if each call were the def's value, every parameter replaced by the argument written for it:

```
priority:
  cast multi_shot
      if $use_multi_shot
      and (spell(multi_shot).cast_time + $move_duration <= $time_to_next_weave or not $melee_weave)
  cast aimed_shot
      if $use_aimed_shot
      and (spell(aimed_shot).cast_time + $move_duration <= $time_to_next_weave or not $melee_weave)
```

A def names a value expression. A `var` is a value the sim reads when the rotation runs, and takes no arguments. A def
is read when the file is read, and takes arguments, which lets one value serve several spells, auras or units. No def
reaches the proto.

**Forms.** `def name(p1, p2, …) = <expr>` declares a def with parameters, each an ident, and `def name = <expr>` (or
`def name() = <expr>`) one with none. `name(a1, a2, …)` and bare `name` call it where a value goes (§6.5): in a row,
a `var` value, a `when` condition, a `with` value and another def's value. A word that names a def reads as the def in
a value position, before the kinds of §8.3, and where a name goes as below; in step position it is `unknown-kind`.

**A def for a name.** A def with no parameters whose value is a name alone, a word of the table, `item.<name>` or
`other.<name>`, also stands for that name where a name goes: the spell of an action (`cast filler`), the subject of a
head (`spell(filler)`, `aura(filler)`, `dot(filler)`) and any other field that takes a name. One line then picks a
spell that every row casting or reading it follows, and a def with parameters may be passed it:

```
def filler = shadow_bolt
def main_dot = corruption
def short(x) = dot(x).remaining_time < 2s

priority:
  cast main_dot if short(main_dot)
  cast filler if spell(filler).cast_time < 3s
```

The name is read as if written at the call, so a diagnostic about it is at the call, as in a value. A def with
parameters written bare where a name goes is `def-arity`, at the call ("fits takes arguments, so it stands for no
name"). A def whose value is not a name gets there what its value gets where a name goes.

**Where.** At top level or under `apl:`, where a rotation `var` line goes. Anywhere else it is `misplaced-def`, at
`def`, and the line is still read. A def may be called anywhere in its file, before or after its line.

**Scope.** A def belongs to its file and to the files that `use` it (§9): a file sees its own defs, those of the files
it uses, and those of the files these use in turn. `use … with` cannot set one. A used file's def reads at the call
like any other, so the variables and names in its value are read in the calling file, and a diagnostic in its value
names the used file: `(in def far, at shared.apl.siml:4:5)`. A def named as one of this file's, or that two used files
declare, is `duplicate-def`, at the `use` line that brings the second. A file that holds such shared defs and
variables needs no row: `apls/weave_common.apl.siml` of the hunter holds the walk times and conditions both weave
rotations read.

**Calls.**

- Arguments are positional, each one term (§4). A named argument (`fits(x=y)`) is `positional-only`, at the argument.
  A call with another number of arguments than the def has parameters, bare `name` having none, is `def-arity`, at
  the call ("fits takes 1 argument").
- A call reads as the def's value with each parameter replaced by its argument, on the syntax tree, before typing
  (§8). In the value, a word equal to a parameter's name is that parameter in every position: a ref (`spell(x)`), an
  operand, an argument, and a unit before a dot (`x.current_health`, `x.aura(a).is_active`). Written bare, a
  parameter shadows a name of the table, a kind or a def of the same text. A unit before a dot takes a unit word
  only, and any other argument there is `wrong-type`, at the argument.
- The value is typed where the call is, so an argument is a ref, a unit or a value as its parameter's position
  needs. Variables, placeholders and names in the value are read at the call, so `$move_duration` in a group's row
  may mean that group's variable.
- An argument stays one node, as if written in parentheses: passing `a or b` for `x` in `x and c` reads as
  `(a or b) and c`, and passing `a and b` reads as `(a and b) and c`. A call keeps the def's value together the same
  way: `$a and fits(x)` reads as `$a and (…)`.
- A diagnostic in the def's value is reported at the call, its message ending in `(in def fits, at L:C)`, where it is
  in the value. One at an argument stays at the argument, its message ending in `(passed to def fits, at L:C)`, where
  the def's name is. A diagnostic of a call in a def's value moves again at each call around it. One that reaches a
  place by several paths of calls, as `def b = a and a` reaches `a`'s value twice, is reported there once, with the
  first path in its message.
- Typing reads a def's value at each call, and only there, so a def that no call reads has only its name and
  parameters checked.
- A file's calls, in a value and where a name goes, expand at most 65536 times, and each call past that is `parse`
  ("the defs expand past 65536 calls").

**Names and errors.**

- A def's name must not be a reserved word (§4 note 6), a head, a unit word, a kind of §8.3, or a name the table holds
  for the file's class (§7.1): `def-shadows`, at the name. Such a def is not read.
- A second def of a name is `duplicate-def`, at its name, and is not read. A parameter written twice is
  `duplicate-parameter`, at the second.
- A def that calls itself, directly or through other defs, is `reference-cycle` at its name. Its calls read as nothing
  and report nothing more. The calls in a used file's def read as this file reads them, so a cycle may pass through
  this file's defs and those of several used files. A used file's def on a cycle that holds none of this file's defs
  is `reference-cycle` at the `use` line that brings it, unless its own file has the cycle, whose error that line
  reports already.
- A def that no row, `var`, `when`, `with` or other def calls is the warning `unused-def`, at its name, in a file
  that writes a row; a file with no row holds its defs for the files that use it. A parameter its
  value never mentions is the warning `unused-parameter`, at the parameter.

`format` keeps every def line and every call (§12.3), and `print` writes neither.

---

## 7. Names

A file holds names only, and **a name means what the name table says it means**. The table is
`sim/siml/names.json`: a Go generator writes it from what the sim registers and the database (§7.4), and it is
committed. Parsing reads the table and the file's class, and nothing else: no game data, item database, registered
spell list or sim state. This means:

- The same text means the same proto in every implementation, checkout and host.
- A committed file keeps its meaning when the sim, the spell data or the database change, since the table never
  changes what a name it holds means (§7.4).

### 7.1 Namespaces and the class

A position (§7.2) reads names in one **namespace**:

| Namespace      | Holds                 | Read per class |
| -------------- | --------------------- | -------------- |
| `spell`        | what the player casts | yes            |
| `aura`         | auras on the player   | yes            |
| `target_aura`  | auras on a target     | yes            |
| `target_spell` | what a target casts   | yes            |
| `pet_aura`     | auras on a pet        | yes            |
| `item`         | items                 | no             |
| `enchant`      | enchant effects       | no             |
| `suffix`       | random suffixes       | no             |

**The class.** The five spell-like namespaces are read with the `Class` that the file's own `spec` line sets
(§8.4). A spell-like name means the class's entry in the namespace, else the **shared** entry: the spells and auras
that gear, consumables and encounters add, which every class reads (§7.4).

- A file with no `spec` line has no class, and reads the shared entries only. So does a file whose `spec` line
  names no spec (`unknown-spec`).
- A file that `use` reads (§9) reads its names with its own `spec` line, and the values of a `with` list are read
  in the using file.

In one namespace a name means one thing for a class: the table never gives a name that a class holds another
meaning among the shared entries.

### 7.2 Positions

| Position                                                                                                                                                     | Namespace                     | JSON                                                     |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------ | ----------------------------- | -------------------------------------------------------- |
| the subject of `cast` and `channel`, the `spell` and `dot` heads, `multidot`, `Cooldown.id`, and every other `ActionID` field this table does not name       | `spell`                       | `{"spellId": id}`, with the entry's `tag` and `rank`     |
| an `ActionID` field named `aura_id` (the `aura` head, `cancel_aura`, …) whose kind's unit (§6.7) is none, `self`, `playerN`, `all_players` or `unknown_unit` | `aura`                        | as above                                                 |
| … whose unit is `target`, `targetN`, `previous_target`, `next_target` or `all_targets`                                                                       | `target_aura`                 | as above                                                 |
| … whose unit is `petN`                                                                                                                                       | `pet_aura`                    | as above                                                 |
| the `ActionID` field of a kind named `boss_spell_<prop>` (the `boss_spell` head)                                                                             | `target_spell`                | as above                                                 |
| `item.<name>` in any `ActionID` position                                                                                                                     | `item`                        | `{"itemId": id}`                                         |
| `other.<short>` in any `ActionID` position                                                                                                                   | the `OtherAction` enum (§8.2) | `{"otherId": "OtherAction…"}` (`other.offensive_potion`) |
| `none` in any `ActionID` position                                                                                                                            |                               | `{}`                                                     |
| `ItemSpec.id`, the item of a slot row, and the item ref fields of `ConsumesSpec` (§8.1 `refs`)                                                               | `item`                        | the item id                                              |
| `ItemSpec.enchant`, `+ <name>` on a slot row                                                                                                                 | `enchant`                     | the effect id                                            |
| `ItemSpec.random_suffix`, the suffix of a slot row                                                                                                           | `suffix`                      | the random suffix id                                     |

- **Units.** An aura's unit is its kind's `UnitReference` field (the first in declaration order), written as the
  prefix (`target.aura(x)`) or as an argument, before or after the aura (`aura_is_active(x, source_unit=target)`);
  a kind with no such field, or with none written, reads `aura`. The prefix counts before an argument, and the
  first argument before a second. A raw unit reads by its `type`: `Pet` is `pet_aura`; `Self`, `Player`,
  `AllPlayers` and `Unknown` are `aura`; any other type is `target_aura`. A unit that does not read (a word that is
  no unit, or raw JSON that is no `UnitReference`) counts as none written, so its own error is the only one.
- **`item.`** An item in an `ActionID` position is always written `item.<name>`: `cast item.goblin_sapper_charge`,
  `selected_potion(item.major_frenzy_potion)`. A name without it never falls back to the items, so a spell that
  takes an item's name later cannot change what a committed line means. `item.` is written nowhere else: in an
  item field it is `wrong-type`.
- **`$x`.** A variable in an `ActionID` position, such as `cast $filler`, is `spell-variable`: `ActionID`'s `raw_id`
  oneof has no member that names a variable, so no spell can be picked in one. A def for a name (§6.10) picks one
  where the file is read.
- **Errors.** A name that the position's namespace does not hold for the file is `unknown-name`, at the name (for
  `item.<name>`, at the name after the dot). The table decides its message, which is written
  exactly so:
    1. When another namespace holds the name for the file, the first in the order of §7.1 says what the name is.
       When that is `item` and the position reads a spell-like namespace, the message is
       `<name> is an item; write item.<name>`; otherwise `<name> is <thing>; <place> takes <thing>`, the second
       thing being the position's: "falcons_call is an enchant; a slot takes an item", "battle_shout is a spell;
       this position takes an aura on a target", "multi_shot is a spell; item. takes an item".
    2. Otherwise it is `no <thing> is named <name>`, the thing without its article; then `; did you mean <names>?`
       with the names of the namespace that the file reads within edit distance 2, where an action casts only those
       the file's spec lets a rotation cast (§7.4), at most three, closest first and then in byte order, joined with
       `, `; and for a spell-like namespace in a file with no class (§7.1),
       `; with no spec, a file reads only the spells and auras every class shares`.

    The things are: a spell, an aura, an aura on a target, a boss spell, a pet aura, an item, an enchant, a random
    suffix. The place is `this position` for a name in an `ActionID` position, `item.` for the name after it,
    `a slot`, `a slot row's suffix` and `+` for the parts of a slot row, and the field's name for any other item,
    enchant or suffix field (`offensive_pot takes an item`).

### 7.3 What a name means

- **Spells and auras.** A spell-like name writes its entry's `ActionID`: `{"spellId": id}`, with its `tag` and
  `rank` when they are not 0.
- **The rank.** The sim finds a spell or an aura by its spell id and tag, whatever the rank; the rank rides on the
  proto only for the site's labels. So an `ActionID` with a `spellId` prints by the name of its spell id and tag
  whatever its rank, and reads back with the rank of the name's entry. §12.4 compares ranks that way.
- **Raw refs.** An `ActionID`, item, enchant or suffix that the table does not name prints as raw JSON for its node:
  the ref itself (``cast `{"spellId":123}` ``), or the slot and a raw `ItemSpec` for a slot row (§12.2). It can only
  come from a site export of something the generator never saw, and it reads back exactly.
- **`raw-ref`.** A committed file under `ui/specs` holds no raw ref. The repository check (§7.4) reports `raw-ref`
  at each raw token whose JSON holds a spell id, an item id, an enchant effect id or a random suffix id. The parser
  never reports it.

### 7.4 The name table

**Form.** `names.json` is one JSON object:

```
{
  "classes": {
    "<class>": {
      "<namespace>": {
        "<name>": {"spellId": <id>, "tag": <tag>, "rank": <rank>, "registeredBy": ["<spec>", …], "icdOnlyBy": ["<spec>", …], "uncastable": true},
        …
      },
      …
    },
    …
  },
  "enchants": {
    "<name>": {"effectId": <effect id>, "slots": ["<slot>", …]},
    …
  },
  "icdOnlyItems": {
    "<item name>": ["<spec>", …],
    …
  },
  "items": {
    "<name>": <item id>,
    …
  },
  "shared": {
    "<namespace>": {
      "<name>": {"spellId": <id>, "tag": <tag>, "rank": <rank>, "registeredBy": ["<spec>", …], "icdOnlyBy": ["<spec>", …], "uncastable": true},
      …
    },
    …
  },
  "suffixes": {
    "<name>": <random suffix id>,
    …
  }
}
```

- `<class>` is a class's short name (§8.2), such as `warrior`. `<namespace>` is one of the five spell-like
  namespaces. `<spec>` is a field name of `Player`'s `spec` oneof (§8.4). `<slot>` is an `ItemSlot` short name
  (§5.3).
- `tag` and `rank` are left out when they are 0.
- `registeredBy` lists the specs whose builds register the entry's spell id and tag in its namespace, in byte order;
  for a class's entry, only that class's specs. It is empty when no build registers it, for a name kept for files
  written before. Parsing never reads it: editors offer the names the file's spec registers, and the registration
  check reads it.
- `uncastable`, only on a `spell` entry and left out when false, is true when no build registers the spell castable
  (`SpellStats.is_castable`, the sim's APL flag, by which the site's rotation editor lists the spells a cast can
  pick): an inner spell, such as a trap's effect or the strike a queue casts. Such a name is a `not-castable`
  warning (§13) where an action casts it, the subject of `cast` and `channel` and the `spell_id` of each other action
  kind that has one (`multidot`, `cast_friendly_spell`, …), and the sim casts it all the same; it reads in every
  other position. The message is
  `<name> is a spell no build lets a rotation cast`, then `; did you mean <names>?` with the castable names of the
  file's spec that `<name>` extends, or that extend it, by `_` and more (`raptor_strike` for `raptor_strike_tag0`),
  at most three. Where an action casts, editors offer and suggest only castable names.
- `icdOnlyBy`, only on an `aura` entry and left out when empty, lists the specs of `registeredBy` whose builds list
  the aura only as an internal cooldown trigger, in byte order. For a file of one of those specs the sim finds it in
  `icd_is_ready`, `internal_cooldown` and `trigger_icd` and in no other aura position, where a name of it is
  `icd-trigger` (§13).
- `icdOnlyItems` names the items whose effect the listed specs' builds list only as an internal cooldown trigger, by
  an item name of `items`, with the specs in byte order. It is written again from the builds on each regeneration. For
  a file of one of those specs, `item.<name>` in an aura position outside `icd_is_ready`, `internal_cooldown` and
  `trigger_icd` is `icd-trigger`, as for an `aura` entry's `icdOnlyBy`.
- `slots` lists, in `ItemSlot` order, the slots of the item types that the entry's enchant rows fit.
- **Layout.** The keys of every object are in byte order, each entry is on a line of its own, and each level is
  indented two spaces. Empty objects are left out, lines end in LF, and the file ends in one newline. So a
  regenerated table differs from the committed one by exactly its new and changed entry lines, and by the comma
  JSON then needs at the end of the line before a new entry that sorts last in its object. The repository's
  formatter would lay the table out otherwise, so `.oxfmtrc.json` leaves it alone.
- **What a loader checks.** It reads keys exactly, a number by its value (`1.0` is the whole number 1), and `null`
  as absent, except for the table itself.
  It refuses a table that breaks any of these:
    - the table, each of its sections, each class and namespace, and each spell-like or enchant entry is an
      object; the table and the entries hold only the keys the form shows; `registeredBy` and `slots` are lists of
      strings; `icdOnlyBy` is a list of strings on an `aura` entry, each a spec of its `registeredBy`; `uncastable`
      is `true` on a `spell` entry; each `icdOnlyItems` entry names an item of `items` and is a non-empty list of field
      names of `Player`'s `spec` oneof;
    - every spell id, tag and rank, item id, effect id and random suffix id is a whole number in the range of
      `int32`;
    - every name matches `[a-z_][a-z0-9_]*` and is no reserved word (§4 note 6), which a file would never read as
      the name, and every id is positive;
    - in a namespace of a class, or of the shared entries, a spell id and tag have one name;
    - a name that a class's namespace and the same shared namespace both hold means one spell id and tag;
    - an item id, and a random suffix id, have one name;
    - every class is a `Class` short name, every namespace under a class or `shared` a spell-like one, every
      slot an `ItemSlot` short name, and every spec of `registeredBy` a field name of `Player`'s `spec` oneof.

**Printing names.** The printer writes:

- for an `ActionID` with a `spellId` in a spell-like namespace, the class's name for its spell id and tag, else
  the shared one, whatever its rank;
- for an `ActionID` with an `itemId` and no tag or rank, `item.` and the item's name; with an `otherId`,
  `other.<short>`; with none of the three, `none`;
- for an item id or a random suffix id, its name;
- for an enchant effect id on an item in slot `s`, of the effect's names whose slots hold `s`, the one with the most
  slots and the first in byte order of those, else the effect's first name in byte order. An effect that a one-hand
  and a two-hand weapon row share prints with the one-hand row's name, which fits both kinds of weapon;
- a raw ref (§7.3) for anything that has no name.

The class is that of the `spec` line the printed file holds (§12.1). A file printed with none, which has no spec,
writes only the shared names, and raw JSON for what only a class names, since that is what it reads back.

**Hosts.** Go reads the committed table with `siml.DefaultNames()`, and any table with `siml.ParseNames`. Only a
build with the `with_db` tag holds the committed table, so `lib.wasm` has none, and without the tag
`DefaultNames` is an error naming it. `Parse`, `Format` and the printers take a table. `ReadFile` reads with the
committed table; without the tag it reads a file that writes no name, and refuses one that does with an error naming
the tag. A parsed `File` lists the names the file writes, each with what it means, its namespace and its slot row
(`Names`), and its raw refs as `raw-ref` diagnostics (`RawRefs`), for the checks below; a file with an error lists
those that read. An editor's `Doc` (`Analyze`) gives the names the file writes with their namespaces, those that
mean nothing included (`Refs`). A ref's path leads into the loader's message (§1); a name outside it, in a section
the file kind does not hold, has none.

**Generating it.** `go run --tags=with_db ./sim/siml/cmd/simlnames -w`, or `make siml-names`, reads the
committed table, the sim and the database, and, when an entry changed, writes the table back, whole, renamed into
place. Builds read the committed table and never regenerate it: `make devserver`, `make vscode-apl` and the site
bundle, whose Vite plugin reads the table, depend on the file alone, so after adding a spell the developer runs
`make siml-names` (or `make siml`, which also checks `ui/specs`) and commits the table with the change. CI's
up-to-date check (below) is what catches a table not committed. The generator:

1. **Builds every spec the sim can build**, with no gear, for every playable race and every talents string that
   puts each talent at rank 1, 2, … up to the highest rank any talent has, each capped at its own. Each runs with
   the spec's options at their defaults; with every number of the options at 1; and with every number at 1 and one
   bool of the options true or one enum at another value. So the warrior's `use_battle_shout` registers the tag-0
   Battle Shout aura, and the hunter's `pet_type` a pet. Every party buff, raid buff, debuff and individual buff is
   on, and a build that fails is retried without them, then without talents. A spec none of whose builds succeeds
   adds no names of its own until the sim builds it, and the report (step 11) counts it.
2. **Collects** a build's player spells into `spell`, its player auras into `aura`, its pets' auras into
   `pet_aura`, and its targets' auras and spells into `target_aura` and `target_spell`. Each is collected by spell
   id and tag, with the highest rank any build registers it with, and whether a build registers it castable
   (`SpellStats.is_castable`). Item and other ids are no spell-like names. What these builds register is the
   class's.
3. **Adds the shared entries.** Each spec is built with its first race, every talent at its highest rank, its
   default options and every buff (or, when that fails, as its first build that succeeded with every buff), and
   one of:
    - an item of the database with `core.HasItemEffect` that the class can use, in the item's first slot; a relic
      only for the class whose relic type it is;
    - an item set of the database, with all its items on, when the class can use them all;
    - an enchant effect with `core.HasEnchantEffect`, on the first item by id that it fits, that has no item effect
      and that the class can use;
    - a consumable with an effect, in `extra_ids`, or as a stone or imbue on the main hand of the first weapon that
      takes it, with the profession it needs;
    - a preset encounter, with its targets' inputs as they are, and again with every bool input on.

    What such a build registers beyond what the spec's class registers is shared. A build that fails is reported
    and skipped.

4. **Reads the database:** every item of its items (`db.json`, then the leftover database), its consumables and its
   item icons; every enchant row; and every random suffix.
5. **Takes game names.** A spell's is the database's spell icon name, else the spell store's
   (`sim/core/spelldata`), else none. An item's is the first name the database gives its id, and its item level is
   the item's (`scaling_options` 0) or the consumable's level, else 0. An enchant row's and a suffix's are their
   names.
6. **Makes base names.** A thing's base name is `identifier(game name)` (§7.6), except:
    - for an enchant, a leading `enchant_` is dropped when a letter follows (`Enchant Cloak - Agility` →
      `cloak_agility`, but `enchant_2h_weapon_agility`);
    - an empty base is `<kind>_<id>`, and one that starts with a digit is `<kind>_<base>`;
    - a reserved word (§4 note 6) is `<base>_<kind>`.

    `<kind>` is `spell` for the spell-like namespaces, else `item`, `enchant` or `suffix`.

7. **Finds what needs a name.**
    - For each class, in `Class` order, and each spell-like namespace: every spell id and tag the class registers
      there that neither the class's nor the shared entries name.
    - Then for each spell-like namespace: every shared spell id and tag that the shared entries do not name.
    - Every item id and every random suffix id that the table does not name.
    - For enchants, the rows of one effect id that fit the same slots are one thing, named by the first such row
      in the database's order. Each that has no entry of its effect id and slots needs one.
8. **Gives new names.** A name is **taken** for a new entry when it is reserved, or means something else where the
   entry is read: for a class's entry, in the class's and the shared namespace; for a shared entry, in the shared
   namespace and every class's; for an item, enchant or suffix, in its namespace. The new things of one base name
   are named together, the bases in byte order and the things of a base in preference order:
    - The first takes the base name when it is not taken.
    - Each other takes the base with the first of its qualifiers that no other thing of the base shares, new or
      already named where the entry is read. For a spell-like entry the qualifiers are its tag (`_external` for
      −1, `_tag<N>` for N ≥ 0, `_tag_m<N>` for any other negative N), then its rank when it has one (`_rank<N>`,
      `_rank_m<N>`); for an item, its item level when it is known (`_ilvl<N>`); enchants and suffixes have none.
      Without one, it takes `_<id>` (the spell id, item id, effect id or suffix id), and the tag qualifier after it
      when another thing of the base has the same spell id.
    - While that name is taken, `_2`, `_3`, … follow it.

    **Preference.** In `spell`, castable first. Then, in every spell-like namespace, tag 0 first, a rank that is no
    class book rank (`spelldata.IsBookRank`, the hand-kept list in `sim/core/spelldata/book_ranks.go`, since the game
    files state a book rank like a trained one), the highest rank, the lowest spell id and the lowest tag. So a plain
    name means the trainer's top rank and a book rank takes a qualifier, unless being castable or tag 0 puts the book
    rank first or the builds register no trained rank of its base. Items: the highest item level, then the lowest id. Enchants: the lowest effect id, then the slots
    in `ItemSlot` order. Suffixes: the lowest id. The classes are named before the shared entries, so what a class
    casts takes its game name before what gear adds.

9. **Keeps every pinned entry.** An entry keeps its name, id, tag, rank and slots whatever the inputs; only
   `registeredBy`, `icdOnlyBy` and `uncastable` are regenerated. An id that no build registers keeps its entry,
   with an empty `registeredBy`. A spell, item, enchant row or suffix that the game files rename keeps its old name;
   the generator adds no alias. The one exception is work in progress: in a git checkout, an entry whose name the table at `HEAD`
   (`git show HEAD:sim/siml/names.json`) does not hold is dropped before step 7 when no build registers it, or, for
   an item, enchant or suffix, the database no longer holds it. So a name added for a spell registered only for a while, or
   for a wrong id, goes once that is undone, while a name not committed yet that still means something stays, and so
   does a repin's. Outside a git checkout, or when `HEAD` holds no table, nothing is dropped.

10. **Is deterministic.** The same inputs give byte-identical output.
11. **Reports** on standard error:
    - what it added: every entry, or a count for each scope when there are more than 50 and `-v` is not given;
    - each name a `-repin` gave another meaning, and each `.siml` file under `ui/specs` that uses it;
    - each pinned spell-like name that no build registers while an id of the same base name is registered
      ("warrior.spell.eye_for_an_eye is pinned to spell 9799; the sim registers spell 25997");
    - each entry dropped as work in progress (step 9);
    - how many pinned names no build registers;
    - how many specs and sources it could not build, and with `-v` each with why.

**Changing what a name means.** `simlnames -w -repin <class>.<namespace>.<name>` (or `shared.<namespace>.<name>`,
`item.<name>`, `enchant.<name>`, `suffix.<name>`) drops that entry, and every entry of the same namespace and scope
whose game name gives the same base name, and generates, so steps 7 and 8 name what the sim and the database
hold now. It prints each dropped name's old and new meaning, and every `.siml` file under `ui/specs` that
means one of them by a name it writes. It is the only way a name changes meaning, and it records itself: for each name whose meaning changed, `-w` adds a line `<name>: was <meaning>,
now <meaning>` (`warrior.spell.overpower: was spell 11585, now spell 11584`) to the repin record,
`sim/siml/gen/repinned.txt`, unless the record holds that line. The record is committed with the table, and its
lines are kept.

**Checks.**

- **The table keeps its entries.** A `with_db` test in `sim/siml/gen` regenerates the table in memory and fails
  when the committed file is not in the generator's form, or when an entry of it comes back with another meaning
  or spec list, or an item of `icdOnlyItems` with other specs, listing each with the command `make siml-names`. The
  names the sim and the database give beyond the committed table are logged, not failed: a database update or a
  class change adds names without changing what is committed, and `make siml-names` adds them.
- **Names keep their meaning.** The test above regenerates from the table it is given, so it passes two names
  that swapped their ids. `go run ./sim/siml/cmd/simlnames -base <commit>` reads the table and the repin record at
  `<commit>`, and fails, listing each, when a name of that table means another id, tag, rank or slots, or nothing, in
  the working tree's table (`registeredBy` is no part of a meaning), unless the working tree's record holds a line
  for the name that `<commit>`'s record does not. CI runs it on every pull request against the base branch, so a
  repin passes only with its record line in the diff.
- **Repository files.** `wowsimcli siml fmt -check` checks each file under `ui/specs`, and a `with_db` test in
  `cmd/wowsimcli` runs it over every committed one. A file has zero diagnostics and no raw refs (`raw-ref`), is a
  fixed point of `format`, and, when it names a spec, sits under that spec's class folder, `ui/specs/<class>/`.
  Every name it uses in a spell-like position is registered by the file's spec (`registeredBy`), or in a file with
  no `spec` line by some spec; every item is in the database (its items, consumables or item icons); every enchant on
  a slot row fits that slot (the entry's `slots` hold it) and that item (a row of its effect fits the item, as the
  sim checks it); and every suffix on a slot row is one of its item's random suffix options.

### 7.5 Variables, placeholders, groups, sequences

**Variables.**

- **Declared exactly.** A variable is declared by its exact proto name: `var ${Melee weave} = true`, or
  `var $x = …` when the name is itself an ident.
- **References.** `${Exact}` refers to that exact name. `$x` refers to:
    1. the declared variable whose name equals `x`, if any;
    2. else the one declared variable whose `identifier(name)` equals `x`.

    The declared variables are the rotation's value variables, and those of the rotations of used files. A
    group's own variables and a group reference's `with` replace a value variable of the same name inside that
    group, and declare none.

- **Errors.** Two declarations folding to `x`, with no exact match, give `ambiguous-variable`. No match gives
  the warning `undefined-variable`, with the JSON name set to `x`.
- **Declared but never assigned.** A name that only a group's variables or a group reference's `with` set,
  with no value variable of that name, also gets `undefined-variable`: the sim drops the reference.

**Placeholders, groups and sequences** use exact names: `%x` or `%{Exact}`, `group "Exact"`,
`sequence("Exact")`.

**Reference warnings.** The reference checks give these warnings:

- `undefined-group`, for a group reference, `group "N"` or `group_reference(...)`, that names no group;
- `undefined-placeholder`, for a `with %p` that names no placeholder of that group;
- `undefined-sequence`, for a sequence name that no `sequence "N" [...]` action, or the `sequence` kind
  written generically, defines in a row of the priority list or a group that is not hidden, in this file or a used
  rotation: the sim finds no sequence that only a hidden or a prepull row holds. Sequence names are the string fields
  named like the subject of the `sequence` head: `sequence("N").is_complete` and `reset_sequence("N")`.

**Cycles.** A cycle through variables, or through group references, is `reference-cycle`. The sim cannot run
one. The sim reads every reference in a variable's value, a group's variable's too, as a value variable's, so a
cycle runs through the value variables of the rotation the files make together (§9), with the values `with` gives
them; a group's `var $x = $x + 1s` reads the value variable `$x` and is no cycle. Each value variable on a cycle is
`reference-cycle` at the `with` variable that sets it, else at this file's `var` line, else, on a cycle that passes
neither, at the `use` line that brings it. Each group that refers to itself through group references is
`reference-cycle` at its name.

### 7.6 `identifier()`

`identifier(s)` processes `s` character by character. ASCII letters are lowercased, ASCII digits are kept, and
`'` (U+0027) and `’` (U+2019) are dropped. Every other character is a separator. Runs of separators become one
`_`, and leading and trailing `_` are trimmed. Examples:

- `Hunter's Mark` → `hunters_mark`
- `Saf-T Ultra Precision Scope` → `saf_t_ultra_precision_scope`
- `Ez-Thro Dynamite II` → `ez_thro_dynamite_ii`

`sim/siml/testdata/ident.txt` pins it.

---

## 8. Forms derived from the protos

Every form in this section comes from the proto descriptors plus the small hand-kept table `sim/siml/forms.json`.
**A new proto field, enum value, APL kind or spec needs no change to this document, the grammar, or the
implementation,** with one exception: a field of a message that §6 writes by hand as a row, a group header or a var
line, which a guard test below refuses until it has a form. A kind with a short form (§6.4, §6.5) prints in its
generic form when it sets a field the short form does not write. Go computes the derived names and forms once, as
the schema, when it first reads a file; the editor's grammar takes its words from the same schema (`wowsimcli siml
grammar`).

### 8.1 What `forms.json` holds

| Key           | Contents in v1                                                                                                                                                                                                                                                                                                                  |
| ------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `keywords`    | the reserved words of §4 note 6, except the heads and unit words, which have keys of their own                                                                                                                                                                                                                                  |
| `heads`       | `aura`, `dot`, `spell`, `boss_spell`, `sequence`, each with its subject's printed type: an `ActionID`, or a `string` for `sequence`                                                                                                                                                                                             |
| `units`       | the unit words of §6.6 with their JSON, and the indexed prefixes `target`, `player`, `pet`                                                                                                                                                                                                                                      |
| `short`       | the kinds the keyword actions of §6.4 and the operators of §6.5 write: `cast_spell`, `channel_spell`, `wait`, `wait_until`, `move`, `move_duration`, `group_reference`, `schedule`, `sequence`; `const`, `and`, `or`, `not`, `cmp`, `math`, `variable_ref`, `variable_placeholder`                                              |
| `sections`    | the table of §5.1: name, category, field sources in order, body kinds                                                                                                                                                                                                                                                           |
| `refs`        | `proto.ItemSpec.id`→item, `proto.ItemSpec.enchant`→enchant, `proto.ItemSpec.random_suffix`→suffix, `proto.ConsumesSpec.*_id`, `*_ids`, `potions` and `conjured_items`→item, `*.stat_type1`, `stat_type2` and `stat_type3`→stat, `proto.ActionID.spell_id` and `item_id`, `proto.Target.id` and `proto.Encounter.zone_id`→number |
| `indexed`     | `proto.UnitStats.stats`→`proto.Stat`, `proto.UnitStats.pseudo_stats`→`proto.PseudoStat`, `proto.Target.stats`→`proto.Stat`                                                                                                                                                                                                      |
| `flag_values` | `proto.TristateEffect`→`TristateEffectRegular`                                                                                                                                                                                                                                                                                  |
| `skip`        | `proto.Player.database`, `*.api_version`, fields of type `proto.UUID`                                                                                                                                                                                                                                                           |

**The Go tests that guard this table:**

- a keyword, head or unit word equal to a kind name fails, unless `short` lists the kind: the keyword is how
  `wait`, `and` or `sequence` is written;
- an `int32`, `uint32` or `int64` field reachable from `IndividualSimSettings` whose name is `id`, ends in `_id`
  or `_ids`, or starts with `stat_type`, and has no `refs` entry, fails;
- a field that no section maps, no top-level line sets and `skip` does not name fails, for every message that
  sections read only part of: `IndividualSimSettings`, `Player`, each spec message, `APLRotation`,
  `SimpleRotation` and `EquipmentSpec`;
- a field that `skip` does not name fails in a message §6 writes by hand, unless its syntax writes it: `APLListItem`
  (`hide`, `notes`, `action`), `APLPrepullAction` (`action`, `do_at_value`, `hide`), `APLAction` (`condition` and
  the `action` oneof), `APLGroup` (`name`, `actions`, `variables`) and `APLValueVariable` (`name`, `value`);
- two fields flattened into one section with the same name fail, checked for every spec, and so do two fields
  of one message;
- an enum whose short names collide, or one of whose short names is a keyword, fails;
- a `refs`, `indexed` or `skip` entry that matches no field fails.

### 8.2 Names

**Field names.** A field is written by its proto name. A name containing an upper-case letter is written as
the snake case of its JSON name. An item ref field may drop a trailing `_id` or `_ids` (§5.2).

**`snake(s)`.** Insert `_` before an upper-case letter that follows a lower-case letter or a digit, and before
an upper-case letter that follows an upper-case letter and precedes a lower-case one, then lowercase:

- `FasterAttackII` → `faster_attack_ii`
- `MainHandAuto` → `main_hand_auto`
- `E20` → `e20`

**Enum short names.** For a value `V` of an enum whose short type name is `E`:

1. If `V` starts with `E` and is longer, strip `E`, and then any `_` that follows it.
2. Take `snake` of the result.
3. If that is empty, starts with a digit, is a keyword, or equals another value's short name in the same
   enum, use `snake(V)` instead. If `snake(V)` is a keyword too, use `snake(E)`, `_` and `snake(V)`.

| Value                                  | Short name                                                        |
| -------------------------------------- | ----------------------------------------------------------------- |
| `RaceOrc`                              | `orc`                                                             |
| `TypeAPL` (enum `APLRotation.Type`)    | `apl`                                                             |
| `ItemSlotMainHand`                     | `main_hand`                                                       |
| `TristateEffectRegular`                | `regular`                                                         |
| `MainHandAuto`                         | `main_hand_auto`                                                  |
| `E20`                                  | `e20`                                                             |
| `PetAttackSpeedNone`                   | `pet_attack_speed_none` (the keyword `none` forces the full name) |
| `OtherActionOffensivePotion`           | `offensive_potion`                                                |
| `SwipeUsage_Never` (enum `SwipeUsage`) | `never`                                                           |
| `None` (enum `FinishingMove`)          | `finishing_move_none` (`none` is a keyword, and so is `snake(V)`) |

Adding a value renames another only when their short names collide. The proto value name is also accepted on
input.

**Kind names.** A kind is named by its oneof field's proto name in `APLAction.action` or `APLValue.value`
(`cast_spell`, `aura_is_active`).

### 8.3 Kind forms

Each kind's form is decided in this order:

1. **Short forms.** The keyword actions of §6.4 and the operators of §6.5 cover their kinds whenever the
   kind's fields fit the short form. `cast_spell` fits with `spell_id` and optionally `target`. `wait` fits
   with `duration`. `cmp` fits with `op`, `lhs` and `rhs` all present.
2. **Accessor.** A value kind whose name is `<head>_<prop>` qualifies when its message has exactly one subject
   field (an `ActionID` for `aura`, `dot`, `spell` and `boss_spell`; a `string` for `sequence`) and at most
   one `UnitReference` field. The longest matching head wins.
3. **Bare word.** A kind with no fields is written bare.
4. **Unit prefix.** A value kind whose only field is a `UnitReference` is written `[u.]kind`.
5. **Plain arguments.** A kind whose only field is repeated takes its items as plain arguments:
   `kind(a, b)`.
6. **Generic.** Every other kind is `kind(<positional>, <field>=<value>, …)`.
    - The **positional field** is the first field in declaration order whose printed type is a ref (an
      `ActionID` included), an enum, a stat, a string, a value, an action or a list (any repeated field but a
      map).
    - Units, numbers and bools are always named.
    - An unnamed argument that is a bare word naming one of the kind's bool fields sets that field to true;
      any other unnamed argument is the positional field's value, or one of the list's items for a kind of
      rule 5. A bool field may also be written by its name alone on a property (§6.7); the printer writes
      `name=true`.
    - A positional argument given to a kind that has no positional field, or a second one, is
      `positional-not-allowed`. The positional field written both unnamed and named is `duplicate-field`.
    - A kind written with no arguments has none of its fields set, whatever its form.

**The generic form is always accepted**, even where a shorter form exists (`aura_is_active(x)`), and the
printer writes the shortest form that reads back to the same JSON (§12.3).

**A kind's field order** is the proto declaration order that protoreflect reports.

### 8.4 Specs

The schema lists each field of `Player`'s `spec` oneof with:

- its message;
- its `Spec` enum value, named `Spec` + the message name;
- its `Class`, the one `Class` value whose short name is the message name in snake case or ends it after a
  `_` (`DpsWarrior` → `warrior`);
- its `Options` message, the type of its `options` field, and the Options' `class_options` field when there is
  one;
- its nested `Rotation` message, which `simple:` writes.

`spec <field>` sets `Player.<field>` to that message, empty unless `settings:` fills its options, and sets
`Player.class`.

---

## 9. Composition: `use`

```
use "../gear_sets/p1.gear.siml"
use "default.apl.siml" with $use_summon_hawk = true
```

- **What it takes.** `use` takes every section of the named file into this one, and the sections it takes in
  turn, with its variables and its defs (§6.10). The used file is read as the kind its own name gives, and its
  names with its own `spec` line (§7.1).
- **Order.** This file's sections come first, then each used file's in the order of the `use` lines, which
  decides the order of the groups and of the variables.
- **The path.** It is read against the folder of the using file, with `/` or `\` between parts and `.` and `..`
  resolved; a path that starts with `/`, `\` or a drive (`C:\`) is read as it is.
- **Conflicts.** A section present in both files, or brought by two used files, is `section-conflict`. Two
  groups conflict when they share a name, and two talents lines conflict too. A `def` writes no section, so a
  file whose rotation part is only defs takes in the used rotation whole. Variables are no section either: those
  of every file go together, and only `apl:` entries such as `type` conflict as `apl:`. A variable outside the
  groups that this file and a used file both declare is `duplicate-variable` at this file's `var` line, and one that
  two used files declare is `duplicate-variable` at the `use` line of the second; `with` sets a used file's
  variable. A file that two `use` lines reach, directly or through the files they use, brings its variables once
  where they have the same value through both, after `with`; with another value, one is `duplicate-variable` as
  above. Its defs come in once (§6.10), and its sections conflict as any two files' do. A def two of the files
  declare is `duplicate-def` (§6.10). The message names where each section was first: this file, or the used file
  that brought it. A section file that uses a section its kind does not hold, or a talents line (§3.3), is
  `section-not-allowed`. Each is reported once per `use` line, listing the sections.
- **Specs.** The two files' `spec` lines must agree (`spec-mismatch`).
- **`with`.** It sets the value of variables the used rotation declares, matched by the rule of §7.5. An
  unknown variable is `with-unknown-variable` and one that matches several is `ambiguous-variable`, each at the
  variable, and `with` on a file with no rotation is `with-not-rotation`, at the first variable of the list. A
  list that sets one variable twice is `duplicate-variable`, at the second.
  Names in `with` values are read in the using file, variables in them are this file's and the used rotation's, and
  a call in them may name a def of this file or of any used file (§6.10).
- **References.** This file's rows may refer to the variables, groups and sequences that a used rotation declares.
- **Errors in the used file.** The used file's first error, in the order of §13, is reported at the `use` line
  with its own code, and its message names the used file's line and column; the other checks of that `use`
  line are not made. A cycle is `use-cycle`, reported in the file whose `use`
  closes it, and a missing file is `use-not-found`.
- **Hosts without file access.** A host that cannot read files, such as the site importer, reports
  `use-refused`. `format` follows `use` where it can read files, for the variables and defs the used files declare,
  and writes this file's text alone; without file access it reads the file alone.
- **Where the codes sit.** Every code of this section but those of `with` is reported at the `use` line's
  path.
- **Exports.** They never write `use`.

---

## 10. Comments, notes and hidden rows

- **`//` comments.** They are file trivia, and can be a full line or trailing. They never reach the proto, so
  a site round trip loses them. `format` and editor splices keep them:
    - a trailing comment on a statement's last line stays at its end, and any other comment inside a statement
      goes above it (§12.1);
    - a full-line comment directly above a statement, with no blank line between, moves with that statement;
    - any other comment keeps its place between the statements around it.
- **`///` notes.** One or more `///` lines directly above a `priority:` or group row, with no blank line
  between, are that row's `notes`, in a `when` body too. One optional space after `///` is not part of the text, and
  neither are the spaces and tabs that end the line. A `///` anywhere else is `misplaced-note`, including above a
  prepull row, a `var` line, a `def` line, a `when` line, a header or an entry.
- **Hidden rows.** `#` before a row hides it, and the row stays in the proto with `"hide": true`.

---

## 11. Literals in full

| Literal                                              | Where                    | Meaning                                                              |
| ---------------------------------------------------- | ------------------------ | -------------------------------------------------------------------- |
| number with a unit (`1.5s`, `400ms`, `30%`, `1m30s`) | values                   | `const` with the text exactly                                        |
| number (`5.00`, `7`, `-1`)                           | values; entries          | `const` text exactly, or the field's number                          |
| `true` / `false`                                     | values; entries          | `const "true"` / `"false"`, or a bool field                          |
| `"…"`                                                | values; entries          | `const` with the string, or a string field                           |
| `none`                                               | action, value, unit, ref | the empty message                                                    |
| raw `` `…` ``                                        | any node                 | that node's JSON, which must type-check against the node (`bad-raw`) |

**Reading raw JSON.** Raw JSON is read as the site reads JSON, in two steps:

1. As `JSON.parse` reads the text: a key written twice keeps its last value, and a number is the double nearest
   it, whose own value an integer field takes (`123456789012345670` reads as `123456789012345664`). A 64-bit
   integer past 2^53 is exact only as a string.
2. As protojson reads that value into the node's message, where a field the schema skips (§5.1) is no field. An
   enum number that names none of the enum's values is kept as that number, as protojson and the site keep it, so
   a proto that holds one prints as raw JSON that reads back.

Beyond protojson's own rules:

- a string that escapes a lone surrogate, and a number past the range of a double, are `bad-raw`;
- a `float` field holds the double rounded to a float32 (§5.2), and a number past the range of a float32 is
  `bad-raw`;
- a number written as a string is exactly a JSON number (`"1e3"`, not `"5,6"`, `" 5"` or `"01"`), or for a
  float `"NaN"`, `"Infinity"` or `"-Infinity"`; an integer field takes it when its value is a whole number;
- a `-0` is the default value of a field with no presence, as it is in a JSON number, and is left out;
- a map key reads as protojson reads it (`"+7"` and `"07"` are both 7), and a key written twice that way is
  `bad-raw`.

---

## 12. Canonical printing

`print(proto)` is a pure function of its input and the name table (§7). `format(text)` re-prints every statement
from its typed value and puts back comments, notes and blank lines. **A committed file must equal
`format(file)`.**

### 12.1 Layout

**Whitespace.**

- LF line endings, and exactly one newline at the end of the file.
- No trailing whitespace, and no tabs.

**Order of the file.**

1. `siml 1`, then `spec`, `talents` and `use` lines, in that order. `spec` is written in each file kind that holds
   it (§3.3); `talents` in `x.siml`, when its category is given (`talents ""` for an empty string) or, with no
   categories given, when the string is not empty.
2. The setup sections, in the §5.1 order: `settings`, `gear`, `item_swap`, `consumables`, `buffs`,
   `party_buffs`, `raid_buffs`, `debuffs`, `encounter`, `sim`.
3. The rotation: `apl` (vars), `simple`, `cooldowns`, `prepull`, `priority`, then the groups in proto order.

**Blank lines.** Exactly one blank line separates the top-level lines, each section, and the block of
top-level `var` lines. Inside a section `format` keeps single blank lines that were in the source, and `print`
writes none.

**What `format` writes.** It re-prints each statement in place, so the sections and the statements in them keep
their order; only `print` orders a file as above.

- **Blocks.** The top-level lines, the implied section's body and each section are blocks. Exactly one blank
  line comes before each block but the first. Inside a block a run of blank lines becomes one, and none is
  kept before the file's first line or before a section's first body line.
- **Comments.** A full-line comment takes the indent of the statement below it, and a comment after the last
  statement takes column 1. A comment loses the white space that ends it: tabs, U+000B, U+000C, U+FEFF,
  `\r`, U+2028, U+2029 and every space separator (Unicode Zs: the space, U+00A0, U+3000, …). The comment on a
  statement's last line stays at the end of its last printed line, after one space; the other comments inside a
  statement are written above it, in order. A header's body line goes on a line of its own, and the comment at the
  end of the header line goes with it, except for a section of entries whose only statement is that body line: it
  stays on the header line when it is one line and the two fit in 100 columns. `format` never takes a body up onto
  its header line.
- **A section file's header.** The header of the implied section (§1) is left out, as `print` leaves it out, and
  the comment at the end of its line goes with its body line, or on a line of its own when it has none. A header
  that comes after another section's header, such as `apl:` after `priority:` in an `x.apl.siml`, stays, since
  without it the lines below would belong to that section.
- **Entries** are re-printed as `print` writes the fields each one sets (§12.2): a run of flags of one line as
  flag lists, the values one line gives a list together, and every other field on a line of its own.
- **Slot rows** are re-printed as §12.2 writes them, with the names they are written with.
- **`when` blocks and defs** are kept, as §12.3 writes them. `print` writes flat rows and each call's value, and
  never writes `when` or `def`, so §12.4 holds as it is.

**Indentation.**

- Section bodies and rows are indented 2 spaces, except in section files, where the implied section's body is
  at column 1 but for the lines under a header of it that stays. An `x.group.siml` writes each group under its
  header (§1), indented as in any other file.
- In an `x.apl.siml` the rotation's `var` lines are at column 1. In `x.siml` they sit under `apl:`.
- Continuation lines are indented 4 more than their statement.
- A `when` body is indented 2 more than its `when` line, and a `def` line as a `var` line in its place.

**When a body goes on the header line.** A section whose body is a single entry line (not a slot row, row or
`var` line) is written on the header line (`buffs: greater_blessing_of_kings`) when that fits in 100 columns.
Otherwise each entry goes on its own line, and a flag list is packed greedily into lines of at most 100
columns.

**`apl:` in a full file.** It is printed when the rotation is present and has value variables, or needs a
`type` entry because its type is not the inferred one (`TypeSimple` with `simple`, else `TypeAPL`), or would
otherwise get no section at all: no `simple`, `prepull`, `priority` or group.

**Which sections a full print writes.** `print(proto, categories)` writes:

- each section of a given category whose target message is present (§5.1);
- each section with at least one field set;
- each section with at least one row.

When a given category gets no section this way, its first section in §5.1 order is written with an empty body,
so the file still holds that category. A section whose category is not given is never written. With no
categories given, every section that the three rules above find is written. A section file writes its own
section only, with no header but a group's; an `x.gear.siml` writes its slot rows and no entry.

### 12.2 Entries and slot rows

**Entries.**

- Fields follow the section's order (§5.1). Zero values are left out, and so are absent messages.
- A run of consecutive present fields that are all true bools or flag-value enums becomes one flag list, in
  declaration order. A flag-value enum at its flag value is written bare, and at another value `name=value`.
- Repeated values are joined with `, `, and a message is written `field(...)` with its fields in declaration
  order: a true bool, or a flag-value enum at its flag value, by its name alone, each message of a list as a
  `field(...)` of its own, and every other field `name=value`.
- An indexed array lists its non-zero entries in index order. One that is all zeros is left out, as §12.4 reads
  it as absent. One with an entry that no value of its enum names, or that is no number, is written
  `field=` and raw JSON.
- A field named `siml`, `spec`, `talents`, `use`, `var` or `def` is written `name=value` on a line of its own (§4
  note 5).
- An entry line longer than 100 columns breaks after each comma between a list's values, and after each comma
  between the top-level arguments of a `field(...)`; each piece goes on a continuation line.
- The printer reads `spec_rotation_json` as raw JSON of the spec's `Rotation` (§11), with keys that name no
  field ignored. A string that does not read so writes no field, as the site reads it as the default rotation.

**Slot rows.**

- Rows follow `ItemSlot` order, and empty slots are left out. Slot rows hold the slots `ItemSlot` names, the 17
  items of §5.3.
- The slot word is padded with spaces to 11 characters, and followed by at least one space.
- An `ItemSpec` that is not an item id with an optional suffix and enchant, or whose item, suffix or enchant has no
  name (§7.4), is written as the slot and raw JSON.

### 12.3 Rows and expressions

**Rows.**

- Each row and each value takes the shortest form of §8.3 that reads back to the same JSON. If none does, the
  generic call is used, then raw JSON for the smallest node that needs it.
- Every printed statement (row, `var` line, entry, slot row) is checked by reading the text back. One that does
  not read back is written as raw JSON: a raw row, a raw entry of the fields it sets, or the slot and a raw
  `ItemSpec`. A spell-like name reads back when its entry's spell id and tag do, whatever the rank (§7.3).
- A key that names no field of the schema, as the skipped fields of §5.1 (`uuid`, `api_version`) do, is never
  printed.

**Expressions.**

- Operators are written with single spaces around them.
- `not` is followed by one space.
- Parentheses appear only where §6.5 needs them, including a right operand at equal precedence (`a - (b - c)`).

**Strings, numbers and raw JSON.**

- A string is written as ECMAScript's `JSON.stringify` writes it: `"` and `\` escaped, `\b \f \n \r \t`
  for those control characters and `\u00xx` (lower-case hex) for the other control characters, `\udxxx` for a
  lone surrogate, and every other character as itself.
- A `const` is written bare when its text is a literal (§2), `true` or `false`, and as a string otherwise.
- A number is the shortest decimal that reads back to it, as ECMAScript's `Number#toString` writes it, with
  any exponent written out (`1e-7` is `0.0000001`); a float's is the shortest that reads back to the same
  float32 (§5.2), and a 64-bit integer is its digits.
- Raw JSON is the node's protojson as §12.4 normalises it, compact, with the keys in proto declaration order,
  strings as above and numbers as `JSON.stringify` writes them. A backtick in it is written `\u0060`. A map's
  keys are written as protojson writes them, the array-index keys first by number (as `JSON.stringify` orders
  them), then the others in byte order (`{"5":1,"10":2,"-1":3,"-2":4}`).
- A field written with no value, such as an entry that sets an enum to its zero value, is written as its
  default: `false`, `0`, `""`, the enum's zero value, or `none`.

**Line breaking.** A line longer than 100 columns, indent included, is broken; nothing else is. A comment at
the end of the line does not count.

- A row breaks before ` if` and before each top-level `and`/`or` of its condition.
- A `var` line, and a `def` line, breaks after `=`, then before each top-level `and`/`or` of its value.
- A `when` line breaks before each top-level `and`/`or` of its condition, its `:` ending the last piece.

Each piece goes on a continuation line. A continuation line may itself exceed 100 columns.

**Variables and placeholders.**

- A variable reference prints as `$name` when the name is an ident that §7.5 reads back as itself, else as
  `$identifier(name)` when that is an ident that reads back as the name, otherwise as `${Exact}`.
- A declaration prints as `var $x` when the exact name is an ident, otherwise as `var ${Exact}`, followed by
  `=` and its value, or by nothing for a variable with no value.
- Placeholders print as `%x` when the name is an ident, otherwise as `%{Exact}`.
- In `${Exact}` and `%{Exact}`, `\`, `}` and a line feed are written `\\`, `\}` and `\n` (§2), and every other
  character as itself.

**`when` blocks and defs.** `format` writes them as follows.

- A `when` line is written `when <condition>:`, or `# when <condition>:` for a hidden block, the condition as a row's
  condition is written.
- Each row of a body is written with its own `#` and its own condition, as they are before the join (§6.9): the
  condition its conjuncts give. A condition written as one conjunct, on a row of a body or a `when` line, that would
  print as a top-level `and` is written in parentheses, so that the join keeps it one value.
- A `def` line is written `def name(p1, p2) = <value>`, or `def name = <value>` for a def with no parameters, and a
  call as it is written. Both write the value or arguments as written, with single spaces around operators and `, `
  between arguments. A row, `var` line or `when` line around a call is written from what it reads as before the call
  is replaced (§6.10).
- Each row is read back against the row it reads as, its block's conditions and the defs' values in place.

**Notes.** They print as `///` lines, one per line of the text, and an empty line as `///` alone. A note that
holds a control character other than the line feed (a tab or `\r` among them), or a line that ends in a space,
makes its row print as a raw row (§6.3).

### 12.4 The exactness contract

For every proto `p` that a site export or a repository file holds, `parse(print(p))` equals `p` after these
normalisations:

- `uuid` and `api_version` fields are cleared;
- `EquipmentSpec.items` and `ItemSwap.items` are padded to 17;
- indexed arrays are padded with zeros to their enum's length, an absent one included;
- `spec_rotation_json` is compared as the protojson of the spec's `Rotation` it holds, an empty string as `{}`, in
  an `APLRotation` too, whose spec is the one its file's `spec` line names;
- an `ActionID` with a `spellId` that a spell-like name prints takes the rank of the name's entry (§7.3).

The protos are then compared with `proto.Equal`.

### 12.5 Editor splices

The rotation panel edits the rotation of an `x.apl.siml` through its JSON, and the text changes only where the
JSON does:

- A statement an edit touches is printed again, with the `//` lines directly above it and its trailing comment
  (§10). A statement that moves takes those lines along. Every other line stays as written.
- A row printed again stays in its `when` blocks while its condition still starts with theirs, and splits them
  otherwise (§6.9).
- A `when` block the edits leave with no row goes. When its rows moved out of it, or it was taken apart, the `//`
  lines directly above its line go above the first row that was in it; when they were removed, the lines go too.
- A def call in a row printed again stays while the edits leave its value as it was (§6.10).
- A def whose parameters or value an edit sets is written again as `format` writes it, and a comment inside it goes
  above it.
- The JSON lists the variables that used files bring after the file's own (§9). An edit of one of them is refused, as
  the used file writes it, and a variable put among them goes after the file's own.
- An edit is refused when it removes or renames a variable that a `with` list of this file reads, and a def rename
  when a file that uses this one calls the def, which F2 in the text renames there too (§9).
- The new text is read back, and a statement that does not read as its JSON is printed again, or as raw JSON,
  until each one does.

---

## 13. Diagnostics

Each diagnostic has a range, a severity, a code and a message. In repository files under `ui/specs`, warnings
are errors.

| Code                                                                                                          | Severity | When                                                                                       |
| ------------------------------------------------------------------------------------------------------------- | -------- | ------------------------------------------------------------------------------------------ |
| `bad-character`                                                                                               | error    | a character that starts no token                                                           |
| `unclosed-string`, `unclosed-raw`, `unclosed-bracket`                                                         | error    | end of line or file inside one                                                             |
| `expected-name`                                                                                               | error    | `$` or `%` not followed by a name ("expected a name after $")                              |
| `missing-version`, `unsupported-version`                                                                      | error    | §3.3                                                                                       |
| `misplaced-statement`                                                                                         | error    | §3.3, and a group file's line before its first header (§1)                                 |
| `indented-header`, `unknown-section`, `section-not-allowed`, `duplicate-section`                              | error    | §3.2                                                                                       |
| `misplaced-note`, `misplaced-var`, `bad-hidden-row`                                                           | error    | §10, §6.1, §6.3                                                                            |
| `misplaced-when`, `misplaced-def`                                                                             | error    | §6.9, §6.10                                                                                |
| `empty-when`                                                                                                  | warning  | §6.9                                                                                       |
| `missing-time`, `unexpected-time`                                                                             | error    | a prepull row without a time, or another row with one                                      |
| `spec-required`, `unknown-spec`, `spec-mismatch`                                                              | error    | §3.3, §9                                                                                   |
| `use-not-found`, `use-cycle`, `use-refused`, `section-conflict`, `with-unknown-variable`, `with-not-rotation` | error    | §9                                                                                         |
| `parse`                                                                                                       | error    | an unexpected token, listing what was expected; §4 note 10                                 |
| `write-not`                                                                                                   | error    | `!` ("write not")                                                                          |
| `write-and`, `write-or`                                                                                       | error    | `&&`/`&`, `\|\|`/`\|` ("write and" / "write or")                                           |
| `write-eq`                                                                                                    | error    | `=` where a comparison is expected ("write == to compare")                                 |
| `chained-comparison`                                                                                          | error    | `a < b < c` ("comparisons do not chain: write a < b and b < c")                            |
| `expected-if`                                                                                                 | error    | tokens after a complete action ("a condition starts with if")                              |
| `args-on-head`                                                                                                | error    | `aura(x, f=v).p` ("arguments go on the property: aura(x).p(f=v)")                          |
| `unknown-kind`, `unknown-property`, `unknown-field`, `duplicate-field`, `positional-not-allowed`              | error    | §8.3, §5.2; each message carries a did-you-mean                                            |
| `wrong-type`, `unknown-enum`, `bad-literal`, `bad-raw`                                                        | error    | a value that does not fit its field                                                        |
| `unknown-slot`, `duplicate-slot`                                                                              | error    | §5.3                                                                                       |
| `unknown-name`                                                                                                | error    | §7.2; the message names where the name is, or a did-you-mean                               |
| `spell-variable`                                                                                              | error    | §7.2; a variable where an `ActionID` goes                                                  |
| `icd-trigger`                                                                                                 | error    | §7.4; an `icdOnlyBy` aura or an `icdOnlyItems` item outside ICD kinds                      |
| `not-castable`                                                                                                | warning  | §7.4; an `uncastable` spell where an action casts it                                       |
| `ambiguous-variable`                                                                                          | error    | §7.5                                                                                       |
| `duplicate-group`, `duplicate-variable`, `reference-cycle`                                                    | error    | §6.1, §7.5, §6.10; a variable two files declare, or a `with` list that sets one twice (§9) |
| `undefined-variable`, `undefined-group`, `undefined-placeholder`, `undefined-sequence`                        | warning  | §7.5                                                                                       |
| `def-shadows`, `duplicate-def`, `duplicate-parameter`, `def-arity`, `positional-only`                         | error    | §6.10; a def two files declare, at the `use` line (§9)                                     |
| `unused-def`, `unused-parameter`                                                                              | warning  | §6.10                                                                                      |
| `raw-ref`                                                                                                     | error    | §7.3; reported by the repository check only, never by the parser                           |

**Recovery.** Each statement is read on its own, so a bad statement costs only itself. A file with any error
produces no proto. The diagnostics are exactly these:

- A statement is read only up to its first token with a lexer diagnostic (§2), and is then left unread. Every
  lexer diagnostic is reported, and a parse error only when it lies in front of that token.
- A statement's first parse error ends it. `missing-version`, `unsupported-version`, `misplaced-statement`,
  `indented-header`, `unknown-section`, `section-not-allowed`, `duplicate-section`, `duplicate-group`,
  `misplaced-var`, `misplaced-def`, `missing-time` and `unexpected-time` do not: the statement is still read.
- A `when` line that does not read still holds its body, whose statements are read. A misplaced one's body is not
  read, as under a header that names no section.
- An error at the end of a statement is reported where its last readable token ends.
- A bracket still open at the end of the file is `unclosed-bracket` at the outermost one.
- A `#` line that does not read as a row, by a parse error or any lexer diagnostic, is `bad-hidden-row` alone.
- Under a header that names no section, statements are not read and notes are not checked; their lexer
  diagnostics are still reported.
- `missing-version` is reported at the first statement, or at 1:1 in a file with none, and `spec-required` at
  the `siml` line, or at 1:1 in a file with none.
- Typing reads every node it can: an error in one argument does not stop its siblings, and a node whose kind,
  property or field is unknown is not read further.
- Diagnostics are ordered by where they start, then by where they end, then by code.

---

## 14. Versions

`siml 1` is the major version. It goes up only when existing text would change meaning: a keyword, a literal
rule, a §8 derivation rule, or a section's mapping. New proto fields, enum values, kinds and specs need no
bump. Reserving a word that the name table cannot hold, as `when` and `def` are (§4 note 6, §7.4 "What a loader
checks"), changes the meaning of no text and needs none either.

A proto rename breaks the files that use the old name loudly, in the parser and in CI.

---

## 15. Conformance corpus

The corpus lives in `sim/siml/testdata/`, and `sim/siml/conformance_test.go` runs it. A case's kind comes from its
double extension. A case with no double extension is a full setup.

```
sim/siml/testdata/
  ident.txt                         game name <TAB> identifier, one pair per line
  names.json                        the fixture name table (§7.4) every case is read and printed with
  tables/refused/<case>.json        a name table that breaks a rule of §7.4 "What a loader checks"
  tables/accepted/<case>.json       a name table that keeps them, at their edges
  parse/<case>.<kind>.siml          input
  parse/<case>.<kind>.json          expected JSON of the loader's message for that kind
  parse/<dir>/main.<kind>.siml      multi-file cases (use); expected JSON in parse/<dir>/main.<kind>.json
  print/<case>.<kind>.json          input proto JSON; an apl case's name starts with the spec's field
  print/<case>.<kind>.siml          expected print output
  print/<case>.categories           (optional) the category letters a full print case is printed with
  format/<case>.<kind>.siml         input
  format/<case>.<kind>.siml.want    expected format output
  errors/<case>.<kind>.siml         input
  errors/<case>.<kind>.siml.diag    expected diagnostics, one "LINE:COL severity code" per line, in order
  <case>.<kind>.siml.raw            (optional, beside a parse/ or errors/ case) its raw-ref diagnostics, as .diag
  errors/<dir>/main.<kind>.siml     multi-file error cases (use), with main.<kind>.siml.diag beside it
  splice/<case>.apl.siml            input of the rotation panel's edits (`splice_test.go`)
  splice/<case>.edits.json          AplEdit list (splice.go); a set with no value removes what its path names
  splice/<case>.apl.siml.want       expected text after the splice
  splice/<dir>/main.apl.siml        multi-file cases (use), with main.edits.json and main.apl.siml.want beside it
  builds/<class>_<spec>_<build>.json each committed ui/specs build loaded over its defaults (`build_db_test.go`)
```

**File access.** A file reads the files it `use`s from the file system, resolved as §9 says, except an error
case directly in `errors/`, which is read as a host without file access (`use-refused`). In a folder of a case
of its own, only `main.<kind>.siml` is a case; the other files are the ones it uses.

**Name tables.** The corpus cases are read and printed with `testdata/names.json`, a small table in the form of
§7.4, laid out as the formatter writes the corpus's JSON, so `go test ./sim/siml/...` needs no `with_db`. The
committed `ui/specs/**/*.siml` files are read with the committed `sim/siml/names.json`, read from disk.

**What the package must do:**

1. Parse every `parse/` case to its expected JSON, with no error.
2. Print every `print/` case to its expected text exactly, and parse that text back to the input; a case with
   categories reads back as a file holding exactly those categories.
3. Format every `format/` case to its `.want` exactly.
4. Produce exactly the listed diagnostics for every `errors/` case, matching line, column, severity and code,
   and for a case with a `.raw` file exactly the raw refs (§7.3) it lists, which a file with an error lists too.
5. For every `.siml` in the corpus and every committed `ui/specs/**/*.siml`, meet each of:
    - `format(format(x)) == format(x)`;
    - committed files are a fixed point of `format`;
    - `parse(print(parse(x))) == parse(x)`;
    - zero diagnostics and no raw refs (§7.3) on committed files.
6. **The schema property test.** For every kind and every section field in the schema, build a sample with
   each field set to a typical value, then check `parse(print(sample)) == sample`, and check that the printed
   form is the one §8.3 or §12.2 derives. A new kind or field is covered the day it is added.
7. Refuse every table in `tables/refused`, and load every table in `tables/accepted` (§7.4).

Item 5 for committed files is the repository check (§7.4), which `go test --tags=with_db ./cmd/wowsimcli/...` runs
over `ui/specs`; item 6 is `property_test.go`. The examples of §16 read with the committed table, with no diagnostic
and no raw ref.

**Minimum coverage:**

- every short form of §6.4 and §6.5;
- every position of §7.2, with `item.` and `other.`, a raw ref, and `unknown-name` for a name another namespace
  holds and for one no namespace holds;
- every unit form;
- every section, including present-but-empty sections, and section files;
- `use` with and without `with`, nested, and in a section file;
- the category rules of a full print;
- notes, comments and hidden rows;
- `when` blocks: each join rule of §6.9 (an `and` in the block and in the row, an `or` and parentheses in the block,
  a row with no `if`, nesting, a hidden block, a raw row), in a group too, and their format;
- defs: with zero, one and two parameters, a parameter in a ref, unit and operand position, an `or` argument, a call
  in a `var` value and in a `when` condition, a def under `apl:`, and their format;
- every literal form;
- each enum short-name rule, including the fallback;
- every diagnostic code that the parser emits.

A test fails when the parse cases leave out a section, a form of statement, a short form or another form of §8.3,
or when the errors cases leave out a diagnostic code of §13 that the package emits.

---

## 16. Examples

These are the repository's files for the hunter, and the arms warrior's rotation as SIML writes it; the warrior's
presets are still JSON. The comment lines were added by hand to show `//`; `print` writes none. Each file reads back
to the proto it prints (§12.4).

### 16.1 `ui/specs/hunter/dps/apls/default.apl.siml` (hunter, default rotation)

```
siml 1
spec hunter
use "weave_common.apl.siml" with $wait_at_melee = max(input_delay, $time_to_weave)

var ${Melee weave} = true
var ${Time to weave} = 250ms
var ${Use Aimed Shot} = false
var ${Use Multi-Shot} = true
var ${Use Arcane Shot} = true
var ${Use Summon Hawk} = false

priority:
  autocast_other_cooldowns if current_time >= 5s and at_range
  cast hunters_mark if not target.aura(hunters_mark).is_active
  cast raptor_strike
  cast strider_kick
  cast mongoose_bite
  cast immolation_trap if in_melee
  cast wing_clip if not spell(strider_kick).is_known or spell(strider_kick).time_to_ready > 1.5s
  cast serpent_sting if not dot(serpent_sting).is_active
  reset_sequence("HHA") if hawk_and_arcane and sequence("HHA").is_complete
  sequence "HHA" [cast summon_hawk, cast summon_hawk, cast arcane_shot] if hawk_and_arcane
  cast summon_hawk if $use_summon_hawk and not $use_arcane_shot
  cast arcane_shot if $use_arcane_shot and not $use_summon_hawk
  cast multi_shot if $use_multi_shot
  cast aimed_shot
      if $use_aimed_shot
      and (spell(aimed_shot).cast_time + $time_to_melee <= auto_time_to_next(main_hand_auto) or not $melee_weave)
  /// Start of the fight, before the first weave: get to my spot without losing an Auto Shot.
  group "Walk in" if self.unit_distance != $shoot_spot and $not_weaving_yet
  group "Weave" if $melee_weave

group "Weave":
  start_attack if in_melee
  wait gcd_time_to_ready if gcd_time_to_ready > 0s and gcd_time_to_ready <= 400ms
  wait kick_wait if kick_soon
  wait min($wait_at_melee - auto_time_since_last(main_hand_auto), max(gcd_time_to_ready, input_delay))
  wait min(auto_time_to_next(main_hand_auto) + input_delay, max(gcd_time_to_ready, input_delay))
      if in_melee
      and auto_time_since_last(main_hand_auto) > $time_to_range
  move $melee_spot if walk_in_now
  move $shoot_spot
  wait min($time_to_walk_in, max(gcd_time_to_ready, input_delay))
      if $time_to_walk_in > 0s
      and $weave_fits
```

Some notes on this file:

- The JSON's `"type": "TypeAPL"` is inferred from the file kind (§6.2).
- `"true"` and `"false"` consts print bare.
- The walk times and the weave's conditions come from `apls/weave_common.apl.siml`, which `apls/clean_weave.apl.siml`
  uses too (§9). It uses `apls/pull.apl.siml`, the prepull rows, and `apls/walk_in.apl.siml`, which the turret
  rotations use as well: the group `"Walk in"` that the row above `"Weave"` calls, and `Shoot spot`, which
  `weave_common` sets with `use … with`.
  `walk_in` in turn takes `in_melee`, `at_range` and `hawk_and_arcane` from `apls/common.apl.siml`. Their variables
  follow this file's in the proto, and `with` sets `Wait at melee` from this file's `$time_to_weave`.
- `$shoot_spot` folds to `Shoot spot`, which a used file declares, while the proto keeps the exact names that
  `ui/specs/hunter/dps/spec.ts` sets.
- `move $melee_spot if walk_in_now` reads the used file's def as its condition, so `print` writes the def's four
  conditions on the row.

The two variable presets are one-line files:

```
siml 1
spec hunter

use "default.apl.siml" with $use_summon_hawk = true
```

This is `apls/default_hawks.apl.siml`. `apls/turret_multi.apl.siml` is
`use "turret.apl.siml" with $use_aimed_shot = false, $use_multi_shot = true`.

### 16.2 `arms.apl.siml` (warrior, arms; the warrior's presets are JSON, and this is how its rotation reads in SIML)

```
siml 1
spec dps_warrior

var ${Delay Check} =
    spell(mortal_strike).time_to_ready >= 1.5s
    and (not action_group_used("Stance Dance ON/OFF") or spell(whirlwind).time_to_ready >= 1.5s)

prepull:
  -4.5s cast berserker_rage
  -4s group "Pre-pull: Stance"
  -3s cast bloodrage
  -3s cast battle_shout
  -0.9s group "Pre-pull: Move"

priority:
  # group "Recklessness ON/OFF"
  # group "Stance Dance ON/OFF"
  group "Hamstring Weave ON/OFF"
  group "Engineering"
  group "CD: Trinkets"
  group "Sunder Armor" if number_targets == 1
  # group "Sunder Armor: Refresh" if number_targets == 1
  cast other.offensive_potion
  group "CD: Racials"
  cast death_wish
  cast recklessness if and(action_group_used("Recklessness ON/OFF"))
  cast sweeping_strikes if and(number_targets > 1)
  cast battle_stance
      if not action_group_used("Stance Dance ON/OFF")
      and aura(battle_stance).is_inactive
  cast mortal_strike if action_group_used("Stance Dance ON/OFF")
  cast rend
      if not is_execute_phase(e20)
      and aura(battle_stance).is_active
      and dot(rend).remaining_time <= 1.5s
  cast overpower if aura(battle_stance).is_active
  cast mortal_strike if not action_group_used("Stance Dance ON/OFF")
  cast spearing_strike if not is_execute_phase(e20) and aura(battle_stance).is_active
  cast berserker_stance
      if action_group_used("Stance Dance ON/OFF")
      and aura(battle_stance).is_active
      and aura(overpower).is_inactive
      and (is_execute_phase(e20) or dot(rend).remaining_time > 1.5s)
  when not is_execute_phase(e20):
    cast slam if spell(slam).cast_time < 1.5s
    cast slam
        if spell(slam).cast_time >= 1.5s
        and auto_time_since_last(main_hand_auto) <= input_delay + 300ms
  cast whirlwind if action_group_used("Stance Dance ON/OFF")
  cast hamstring
      if action_group_used("Hamstring Weave ON/OFF")
      and aura(windfury_totem_proc).is_known
      and not is_execute_phase(e20)
      and aura(windfury_totem_buff).is_inactive
  cast execute if is_execute_phase(e20)
  cast battle_stance
      if action_group_used("Stance Dance ON/OFF")
      and not is_execute_phase(e20)
      and aura(berserker_stance).is_active
      and (aura(overpower).is_active(include_reaction_time=true) or dot(rend).remaining_time <= 1.5s)
  cast cleave if number_targets > 1 and or(current_rage >= 40)
  cast heroic_strike if not is_execute_phase(e20) and current_rage >= 90
  cast bloodrage if current_rage < 90
  cast battle_shout
      if $delay_check
      and not aura(battle_shout_external).is_known
      and aura(battle_shout).remaining_time <= 10s
  # cast recklessness if false
  # cast retaliation if false
  # cast shield_wall if false

group "Pre-pull: Stance":
  cast battle_stance if not action_group_used("Stance Dance ON/OFF")

group "Pre-pull: Move":
  cast charge if aura(battle_stance).is_active
  cast intercept if aura(berserker_stance).is_active

group "Recklessness ON/OFF":
  none

group "Stance Dance ON/OFF":
  none

group "Hamstring Weave ON/OFF":
  none

group "Engineering":
  cast item.goblin_sapper_charge
  cast item.crystal_charge
  cast item.ez_thro_dynamite_ii

group "CD: Racials":
  cast blood_fury
  cast berserking

group "Sunder Armor":
  cast sunder_armor if target.aura(sunder_armor).num_stacks(include_reaction_time=true) < 5

group "Sunder Armor: Refresh":
  cast sunder_armor if $delay_check and target.aura(sunder_armor).remaining_time <= 10s

group "CD: Trinkets":
  cast_all_stat_buff_cooldowns(stat_type2=attack_power, stat_type3=haste_rating)
  cast_all_stat_buff_cooldowns(crit_rating, stat_type2=armor_penetration, stat_type3=-1)
  cast item.badge_of_the_swarmguard
  cast item.slayers_crest
  cast item.kiss_of_the_spider
  cast jom_gabbar
  cast item.shard_of_the_fallen_star
  cast item.everlook_pathcarver
```

Some notes on this file:

- **The two Overpowers** are 11585, the cast, and 1282733, the aura that tells it is usable. They share a game
  name and sit in different namespaces (§7.1), so both are `overpower`: `cast overpower` is 11585 and
  `aura(overpower)` is 1282733.
- **The tagged spells.** `heroic_strike`, `cleave` and `whirlwind` are the tag-1 spells, the queued or main-hand
  versions the sim casts from the APL, since those are castable (§7.4). A bare `cast heroic_strike` therefore
  queues the trainer's rank, and the tag-0 Heroic Strikes are `heroic_strike_rank8` and `heroic_strike_rank9`.
- **The Windfury Totem auras** the warrior reads are all external (tag −1) and share a game name, so the
  generator would tell them apart only by id. Their pinned names say what each does instead:
  `windfury_totem_proc` is 10612, the party aura that rolls the proc; `windfury_totem_buff` is 10610, the attack
  power a proc grants; `windfury_totem_aura` is 10614, which the sim holds while the party's totem stands.
- **Items and spells.** The trinkets and engineering items are cast with `item.`. Jom Gabbar's on-use is the
  spell 29602, a shared name (§7.1) that every class reads.
- **Strength.** In `cast_all_stat_buff_cooldowns(stat_type2=…)`, `stat_type1` is strength, `Stat` 0. That is
  proto3's zero value, so it is not printed.
- **One-operand `and` and `or`.** `and(…)`, `or(…)` and the stat-type form read back to the rotation's proto
  exactly. Unwrapping them changes the proto, which only the `.results` files can show is safe.
- **The `when` block.** The two Slam rows read with `not is_execute_phase(e20)` first in their `and` (§6.9), as the
  flat rows wrote it.

### 16.3 `ui/specs/hunter/dps/gear_sets/p1.gear.siml` (hunter, P1 gear)

```
siml 1

// The armor both playstyles share, with the two-hander to weave with.
head       beaststalkers_cap
neck       maelstroms_tendril + necklace_agility
shoulder   beaststalkers_mantle
back       shifting_cloak + cloak_lesser_agility
chest      beaststalkers_tunic + chest_greater_stats
wrist      beaststalkers_bindings + bracer_agility_7656
hands      beaststalkers_gloves + gloves_superior_agility
waist      beaststalkers_belt
legs       beaststalkers_pants + forceful_rugged_armor_kit
feet       beaststalkers_boots + boots_greater_agility
finger1    cutthroats_signet
finger2    selenium_loop of_agility
trinket1   adaptive_combat_assistant
trinket2   rune_of_the_guard_captain
main_hand  legionite_glaive + enchant_2h_weapon_agility
ranged     fahrads_reloading_repeater + biznicks_247x128_accurascope
```

Some notes on this file:

- **The empty off hand** is left out, and the parse still yields 17 items.
- **Effect 904** has three rows, which fit the neck, the wrist and the feet, so it has three names:
  `necklace_agility`, `bracer_agility` and `boots_agility`. On the neck it prints as `necklace_agility`, and
  each of the three reads as 904.
- **Effect 7656** has two rows that fit the wrist, which share one name, that of its first row. 904 is the lower
  effect id and took `bracer_agility`, so 7656 is `bracer_agility_7656`.

### 16.4 The hunter defaults as one full `.siml`

This is `ui/specs/hunter/dps/defaults.siml`, which `ui/specs/hunter/dps/spec.ts` takes its `defaults` from, as one
self-contained file: the committed file has `use "gear_sets/p1.gear.siml"` under its talents line in place of the
`gear:` section. A site export of the same setup is its printed form (§12), with the sections and entries in the
printer's order (`resource_pot` after the elixirs, `encounter:` before `sim:`, `pull_range` last), and also:

- it adds the empty `cooldowns:` section, because the site writes `Player.cooldowns` with every rotation
  (`Player.toProto` in `ui/sim/player/player.ts`);
- it writes `simple.cooldowns` as the entry `cooldowns()`;
- it adds the UI settings in `sim:` when that category is selected.

```
siml 1
spec hunter
talents "-00530501114-550200030050220151"

settings:
  race orc
  profession1 engineering
  profession2 blacksmithing
  reaction_time_ms 250
  distance_from_target 12
  quiver_bonus ancient_sinew_wrapped_lamina
  ammo thorium_headed_arrow
  pet_uptime 1
  cobra_reflexes
  pet_attack_speed faster_attack_ii

gear:
  head       beaststalkers_cap
  neck       maelstroms_tendril + necklace_agility
  shoulder   beaststalkers_mantle
  back       shifting_cloak + cloak_lesser_agility
  chest      beaststalkers_tunic + chest_greater_stats
  wrist      beaststalkers_bindings + bracer_agility_7656
  hands      beaststalkers_gloves + gloves_superior_agility
  waist      beaststalkers_belt
  legs       beaststalkers_pants + forceful_rugged_armor_kit
  feet       beaststalkers_boots + boots_greater_agility
  finger1    cutthroats_signet
  finger2    selenium_loop of_agility
  trinket1   adaptive_combat_assistant
  trinket2   rune_of_the_guard_captain
  main_hand  legionite_glaive + enchant_2h_weapon_agility
  ranged     fahrads_reloading_repeater + biznicks_247x128_accurascope

consumables:
  flask flask_of_natural_aggression
  food swiftstrike_steak
  conjured demonic_rune
  explosive ez_thro_dynamite_ii
  mh_imbue elemental_sharpening_stone
  extra goblin_sapper_charge,
      scroll_of_strength_iv,
      juju_might,
      dragonbreath_chili,
      ground_scorpok_assay
  elixir elixir_of_the_mongoose, greater_mageblood_elixir
  resource_pot major_mana_potion
  offensive_pot major_frenzy_potion

buffs:
  greater_blessing_of_kings
  greater_blessing_of_might 1
  greater_blessing_of_wisdom 1

raid_buffs:
  arcane_brilliance 1
  prayer_of_spirit
  gift_of_the_wild 1
  strength_of_earth_totem 1
  grace_of_air_totem 1
  windfury_totem
  battle_shout 1
  leader_of_the_pack

debuffs:
  judgement_of_light, judgement_of_wisdom, curse_of_elements, curse_of_recklessness, faerie_fire
  expose_armor, sunder_armor, gift_of_arthas, insect_swarm

sim: iterations 25000

simple:
  time_to_weave 250
  pull_range 25
  use_multi, use_arcane
  playstyle weave

encounter:
  duration 180
  duration_variation 5
  execute_proportion_20 0.2
  execute_proportion_35 0.35
  targets(id=31146,
      name="Raid Target",
      level=63,
      mob_type=mechanical,
      stats(attack_power=320, block_value=54, armor=3731, health=1428900),
      min_base_damage=2601,
      damage_spread=0.5,
      swing_speed=2,
      parry_haste,
      can_crush)
```

Some notes on this file:

- **Size.** It is 91 lines and 2.4 KB.
- **The rotation** is `TypeSimple`, inferred from `simple:`. The section encodes
  `{"timeToWeave":250,"useMulti":true,"useArcane":true,"playstyle":"Weave","pullRange":25}` into
  `spec_rotation_json` (§5.4).
- **Left-out fields.** `pet_type pet_none` is the zero value and is not printed. `class` comes from
  `spec hunter`.
- **Buff order.** The fields of `raid_buffs` follow declaration order in `proto/buffs.proto`. `battle_shout` is
  a number, the raid's rank (1 is Rank 6), so it takes a line of its own. So does each rank select, whose number is 0 for
  off and 1 for the rank no class book teaches: `arcane_brilliance`, `gift_of_the_wild`, `prayer_of_fortitude`,
  `prayer_of_shadow_protection`, `strength_of_earth_totem` and `grace_of_air_totem` here, and
  `greater_blessing_of_might` and `greater_blessing_of_wisdom` in `buffs`. A bare name is an error for these.
