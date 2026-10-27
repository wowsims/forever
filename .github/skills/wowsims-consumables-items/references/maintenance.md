# Keeping wowsims-consumables-items true

Re-derive a claim before repeating it:

- the pipeline order: `sed -n '/RouteConsumable(/,/GenerateItemEffectRandomPropPoints/p' tools/database/gen_db/main.go`,
  and `grep -n WriteBinaryAndJson tools/database/gen_db/main.go` for when the files are written
- what loads the database: `grep -n 'go:embed' assets/database/loader.go` and `grep -n READ_JSON ui/sim/proto/database.ts`
- the request's extra items: `grep -n PARTY_BUFF_ITEM_IDS ui/sim/proto/buff_items.ts`
- one effect per item, the aura lists: `grep -n 'panic(\|func implementedByHand\|AuraKinds = ' sim/core/item_effects.go tools/database/gen_effects.go`
- the enchant row accessors: `grep -n 'func GetEnchant\|func FittingEnchantRow' sim/core/database.go`

| You changed…                                                        | Update                                                          |
| ------------------------------------------------------------------- | --------------------------------------------------------------- |
| a step or its order in `tools/database/gen_db/main.go`              | "The pipeline"                                                  |
| routing, shipping, groups, unparsed stats or potion classes         | `docs/consumables.md`, the matching section                     |
| the item or enchant registries, `implementedByHand` or an aura list | "Hand-written effects"                                          |
| the enchant row accessors, the suffix tables or the set rows read   | `references/item-data.md`                                       |
| what the UI adds to a request                                       | "The request database", `docs/consumables.md` §The browser path |
| a ruling's code or its number                                       | `references/rulings.md`                                         |

Then run the shared path check:

```
sh .github/skills/check_paths.sh .github/skills/wowsims-consumables-items
```

After any edit to this skill, run `npm run skills:install` and commit the refreshed `skills-lock.json` with the change: the agents read the installed copy, and the lockfile records the skill's hash.
