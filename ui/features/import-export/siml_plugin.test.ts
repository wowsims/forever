// tools/vite/siml.mts: a .siml import is the JSON of the message its file kind loads into, and a file that does not
// read fails the import at its file, line and column. The plugin runs `wowsimcli siml serve`, $WOWSIMCLI or else
// `go run ./cmd/wowsimcli`.
import { readFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

import { describe, expect, it } from 'vitest';

// A module of tools/, which TypeScript finds by its extension only.
// oxlint-disable-next-line import/extensions
import { simlPlugin } from '../../../tools/vite/siml.mjs';

const ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../../..');
const SPECS = path.join(ROOT, 'ui/specs');
const TESTDATA = path.join(ROOT, 'ui/features/import-export/testdata');
const CORPUS = path.join(ROOT, 'sim/siml/testdata/parse');

type JsonObject = Record<string, unknown>;

// The plugin's transform, called as Vite calls it, with what it watches and warns of written down.
const transform = async (text: string, id: string) => {
	const watched: string[] = [];
	const warnings: string[] = [];
	const context = {
		addWatchFile: (file: string) => watched.push(file),
		warn: (message: string) => warnings.push(message),
		error: (error: string | { message: string }) => {
			throw new Error(typeof error === 'string' ? error : error.message);
		},
	};
	const hook = simlPlugin().transform as unknown as (this: typeof context, code: string, id: string) => Promise<{ code: string; moduleType: string } | null>;
	const result = await hook.call(context, text, id);
	const exported = result ? /^export default (.*);\n$/s.exec(result.code)?.[1] : undefined;
	return { json: exported === undefined ? undefined : (JSON.parse(exported) as JsonObject), result, watched, warnings };
};

// The ActionID the name table gives a hunter's name of `kind`, without the specs that register it.
const NAMES = JSON.parse(readFileSync(path.join(ROOT, 'sim/siml/names.json'), 'utf8')) as {
	classes: { hunter: Record<string, Record<string, JsonObject>> };
};
const hunterId = (kind: string, name: string): JsonObject => {
	const { registeredBy: _, ...id } = NAMES.classes.hunter[kind][name];
	return id;
};

const HUNTERS_MARK = hunterId('spell', 'hunters_mark');
const ROTATION = {
	type: 'TypeAPL',
	prepullActions: [{ action: { castSpell: { spellId: HUNTERS_MARK } }, doAtValue: { const: { val: '-22s' } } }],
	priorityList: [
		{
			action: {
				condition: { not: { val: { auraIsActive: { sourceUnit: { type: 'CurrentTarget' }, auraId: hunterId('target_aura', 'hunters_mark') } } } },
				castSpell: { spellId: HUNTERS_MARK },
			},
		},
		{ action: { condition: { variableRef: { name: 'Use Summon Hawk' } }, castSpell: { spellId: hunterId('spell', 'summon_hawk') } } },
		{ action: { castSpell: { spellId: hunterId('spell', 'multi_shot') } } },
	],
	valueVariables: [{ name: 'Use Summon Hawk', value: { const: { val: 'false' } } }],
};

describe('a .siml import', () => {
	it('is the APLRotation JSON of an x.apl.siml', async () => {
		expect((await import('./testdata/rotation.apl.siml')).default).toEqual(ROTATION);
	});

	it('reads the files a `use` line names, with the values its `with` gives', async () => {
		expect((await import('./testdata/hawks.apl.siml')).default).toEqual({
			...ROTATION,
			valueVariables: [{ name: 'Use Summon Hawk', value: { const: { val: 'true' } } }],
		});
	});

	it('takes the rotation a build uses from the rotation file, so the two imports share it', async () => {
		const build = (await import('./testdata/build.siml')).default;
		expect(build.player.talentsString).toBe('-0053');
		expect(build.player.rotation).toBe((await import('./testdata/rotation.apl.siml')).default);
	});

	it('fails at the file, line and column of what does not read, with the did-you-mean', async () => {
		await expect(import('./testdata/broken.apl.siml')).rejects.toThrow(
			'ui/features/import-export/testdata/broken.apl.siml:5:8: error unknown-name: no spell is named hunters_mrak; did you mean hunters_mark?',
		);
	});
});

describe('the transform', () => {
	it('gives JavaScript, and watches the name table and every used file', async () => {
		const file = path.join(TESTDATA, 'hawks.apl.siml');
		const { result, watched } = await transform(readFileSync(file, 'utf8'), file);
		expect(result?.moduleType).toBe('js');
		expect(watched).toEqual(expect.arrayContaining([path.join(ROOT, 'sim/siml/names.json'), path.join(TESTDATA, 'rotation.apl.siml')]));
	});

	it('leaves an id with a query to the plugin that owns it', async () => {
		expect((await transform('export default "siml 1"', path.join(TESTDATA, 'rotation.apl.siml?raw'))).result).toBeNull();
	});

	const undefinedGroup = 'siml 1\nspec hunter\n\npriority:\n  group "Nowhere"\n';
	const rawRef = 'siml 1\nspec hunter\n\npriority:\n  cast `{"spellId":2643}`\n';

	it('passes a warning outside ui/specs on to Vite', async () => {
		const { json, warnings } = await transform(undefinedGroup, path.join(TESTDATA, 'nowhere.apl.siml'));
		expect(json).toEqual({ type: 'TypeAPL', priorityList: [{ action: { groupReference: { groupName: 'Nowhere' } } }] });
		expect(warnings).toEqual([expect.stringMatching(/^ui\/features\/import-export\/testdata\/nowhere\.apl\.siml:5:9: undefined-group: /)]);
	});

	it('fails a warning under ui/specs, where warnings are errors', async () => {
		await expect(transform(undefinedGroup, path.join(SPECS, 'hunter/dps/apls/nowhere.apl.siml'))).rejects.toThrow(
			/^ui\/specs\/hunter\/dps\/apls\/nowhere\.apl\.siml:5:9: warning undefined-group: /,
		);
	});

	it('fails a raw ref under ui/specs, which writes names only, and reads it anywhere else', async () => {
		await expect(transform(rawRef, path.join(SPECS, 'hunter/dps/apls/raw.apl.siml'))).rejects.toThrow(
			/^ui\/specs\/hunter\/dps\/apls\/raw\.apl\.siml:5:8: error raw-ref: /,
		);
		expect((await transform(rawRef, path.join(TESTDATA, 'raw.apl.siml'))).json).toEqual({
			type: 'TypeAPL',
			priorityList: [{ action: { castSpell: { spellId: { spellId: 2643 } } } }],
		});
	});
});

// A rotation, gear set and full setup of the Go reader's corpus import as the JSON file beside each.
describe('the corpus of sim/siml', () => {
	for (const file of ['hunter_default.apl.siml', 'hunter_p1.gear.siml', 'hunter_defaults.siml']) {
		it(`imports ${file} as its JSON file`, async () => {
			const at = path.join(CORPUS, file);
			const want = JSON.parse(readFileSync(at.replace(/\.siml$/, '.json'), 'utf8')) as JsonObject;
			expect((await transform(readFileSync(at, 'utf8'), at)).json).toEqual(want);
		});
	}
});
