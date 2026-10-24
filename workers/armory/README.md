# Armory Worker

A Cloudflare Worker that looks up a character on the Blizzard Profile API for the sim's Armory importer. It holds
the app's Battle.net credentials, so the browser never sees them, and returns a trimmed character
(`ArmoryCharacter` in `src/character.ts`).

It runs only locally for now. The importer is dev-only until Blizzard publishes the WoW Forever Profile API
namespace.

## Setup

1. Create an API client at <https://develop.battle.net/access/clients>.
2. `cd workers/armory && npm install`
3. `cp .dev.vars.example .dev.vars` and fill in `BLIZZARD_CLIENT_ID` and `BLIZZARD_CLIENT_SECRET`. `.dev.vars` is
   git-ignored.
4. `npm run dev` serves the Worker on <http://localhost:8787>, which is where the sim's dev server looks for it.

## Request

```
GET /character/<region>/<realm>/<name>
```

- `region`: `us`, `eu`, `kr` or `tw`.
- `realm`: the realm name or slug, e.g. `ClassicBetaPvE2`.
- `name`: `First-Last` or `First Last` for a Forever character, or a single name for a Classic one.

The Worker answers 400 for a malformed request, 404 when the character is not found and 502 when the Battle.net
call fails.

## Configuration

`vars` in `wrangler.jsonc` (override one locally in `.dev.vars`):

| Var                        | What it sets                                                                                              |
| -------------------------- | --------------------------------------------------------------------------------------------------------- |
| `PROFILE_NAMESPACE_PREFIX` | The Profile API namespace without its region. `profile-classic1x` (Classic Era) until Forever's is known. |
| `NAME_PART_SEPARATOR`      | How a two-part name is joined in the profile path: `-`, a space or nothing. Unknown until launch.         |

## Tests

The tests run in the repo's vitest suite (`npm run test:unit` at the root). `npm run type-check` here checks the
Worker against the Workers runtime types.
