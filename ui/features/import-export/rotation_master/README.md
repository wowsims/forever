# Rotation Master import / export

Converts between a [Rotation Master](https://rotationmaster.neuromancy.net) rotation (the World of Warcraft addon's
priority list, shared as a text export) and this simulator's APL (Action Priority List). Both directions say, item by
item, how faithful the conversion is. Nothing here touches anything but the rotation.

| File                                                          | What it is                                                                                                                                                                                                          |
| ------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `codec.ts`                                                    | The export string: Base64 of a LibCompress stream (store, LZW, Huffman) holding an AceSerializer value, optionally led by `RM1:`. Bounded reader and writer, no dependencies, nothing evaluated.                    |
| `to_apl.ts`                                                   | Rotation Master rotation -> APL (`rotationToApl`). Branches are flattened, Stop becomes a negated condition on what follows, the Cooldowns list goes first, steps that apply out of combat become pre-pull actions. |
| `from_apl.ts`                                                 | APL -> Rotation Master steps (`aplToRotation`). Groups and variables are expanded in place; a step that is not an exact copy is disabled.                                                                           |
| `index.ts`                                                    | The two entry points, `importRotationMaster(text, context)` and `exportRotationMaster(apl, context)`, plus the words for errors and warnings.                                                                       |
| `report.ts`                                                   | Fidelity levels, the reason codes and their English sentences, rank ladders.                                                                                                                                        |
| `mapping.json`                                                | The rule table both converters interpret. **Copied from the Rotation Master website** (`wowsims/mapping.json`); it is the single description of which Rotation Master condition is which APL value.                 |
| `rm_support.json`, `rm_specs.json`, `rm_condition_names.json` | Generated from the website's tables: the conditions the addon refuses per edition, the specialization ids and talent-tree numbers, the names the editor shows.                                                      |
| `rotation_master.test.ts`, `legacy_exports.ts`                | The tests, and real addon exports (all three compression types).                                                                                                                                                    |

## Fidelity

- **exact**: the same behaviour.
- **approximate**: a close substitute (the simulator's number of targets becomes the enemies you have threat on).
- **lossy**: part of the original was left out (a condition on the fight length; the seconds before the pull). Importing into
  Rotation Master such a step arrives _disabled_; importing into the simulator it arrives enabled with a note, or hidden with
  the "hide steps that lost something" option.
- **unsupported**: nothing to convert (waiting, simulator-only actions, pet actions); it is listed and skipped.

## What it does not do

Rotation Master has no timeline, no fight length, no cast times and no simulator state, so conditions on those are
reported and left out. The simulator has no item sets, pet actions or Assisted Combat, so those steps are reported and
skipped. The seconds of a pre-pull action are assumed (one second apart, ending one second before the pull).

## Keeping it true

`mapping.json` and the three generated tables are the website's. To refresh them, copy `wowsims/mapping.json` from
`rotationmaster-web`, and regenerate the others with the website's tooling; the tests that compare this port with the
website's converters (on every preset rotation of this repository) are kept in the website repository, because they
need its code.
