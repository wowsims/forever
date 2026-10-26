import { existsSync, readdirSync, readFileSync, statSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

import { describe, expect, it } from 'vitest';

import {
	aplToRotation,
	type ClassKey,
	decodeRmExport,
	describeError,
	describeReason,
	describeReasons,
	type Edition,
	encodeRmExport,
	exportRotationMaster,
	exportRows,
	importRotationMaster,
	importRows,
	ladderIndex,
	laddersOf,
	type Reason,
	type RmValue,
	type SimContext,
	type SimRepo,
	specInfo,
	specSlug,
	stripExportWhitespace,
	wrapExport,
} from './index';
import { LEGACY_EXPORTS } from './legacy_exports';

const LADDERS = ladderIndex([
	[
		[133, 1],
		[143, 2],
		[145, 3],
		[3140, 4],
		[8400, 5],
		[8401, 6],
		[8402, 7],
		[10148, 8],
		[10149, 9],
		[10150, 10],
		[10151, 11],
		[25306, 12],
	],
]);
const MAGE: SimContext = { repo: 'classic', edition: 'era', cls: 'MAGE', spec: 'SpecFireMage', actionIdHasRank: true, ladders: LADDERS };
const exportString = (value: RmValue): string => {
	const out = encodeRmExport(value);
	if (!out.ok) throw new Error(out.reason);
	return out.text;
};

describe('the export string', () => {
	it('reads all three LibCompress stream types of a real addon export to the same data', () => {
		const store = decodeRmExport(LEGACY_EXPORTS.store);
		const lzw = decodeRmExport(LEGACY_EXPORTS.lzw);
		const huffman = decodeRmExport(LEGACY_EXPORTS.huffman);
		expect(store.ok && lzw.ok && huffman.ok).toBe(true);
		if (!store.ok || !lzw.ok || !huffman.ok) return;
		expect(lzw.value).toEqual(store.value);
		expect(huffman.value).toEqual(store.value);
		expect(store.envelope).toBe(false);
		const body = store.value as { rotation: Array<{ type: string; action: number; id: string }>; cooldowns: unknown[] };
		expect(body.rotation).toHaveLength(12);
		expect(body.rotation[0]).toMatchObject({ type: 'spell', action: 404, id: '00000000-0000-4000-8000-000000000001' });
	});
	it('writes what it reads, with and without compression, UTF-8 names included', () => {
		const value: RmValue = {
			name: 'Feuer — 雪',
			rotation: [
				{ id: 'a', type: 'spell', action: 133, conditions: { type: 'AND', value: [{ type: 'BUFF', unit: 'player', spell: 1, anyrank: true }] } },
			],
			cooldowns: [],
			priority: 0.5,
		};
		for (const compress of [false, true]) {
			const out = encodeRmExport(value, { compress });
			expect(out.ok).toBe(true);
			if (out.ok) expect(decodeRmExport(out.text)).toEqual({ ok: true, value, envelope: false });
		}
	});
	it('ignores the wrapping chat and mail add, and refuses what is not an export', () => {
		const text = exportString({ rotation: [], cooldowns: [] });
		expect(decodeRmExport(`  ${wrapExport(text, 7)}\n​`).ok).toBe(true);
		expect(stripExportWhitespace(' a b\n')).toBe('ab');
		for (const bad of ['', 'hello world', 'RM2:AAAA', 'AAAA']) expect(decodeRmExport(bad).ok).toBe(false);
	});
	it('marks an envelope with RM1:', () => {
		const out = exportString({ format: 'RotationMaster', schemaVersion: 1, body: { rotation: [], cooldowns: [] } });
		expect(out.startsWith('RM1:')).toBe(true);
		expect(decodeRmExport(out)).toMatchObject({ ok: true, envelope: true });
	});
});

const BODY: RmValue = {
	name: 'Fire mage',
	rotation: [
		{ id: 'pull', type: 'spell', action: 133, ranked: false, conditions: { type: 'NOT', value: { type: 'COMBAT', unit: 'player' } } },
		{ id: 'comb', type: 'spell', action: 11129, conditions: { type: 'SPELL_AVAIL', spell: 11129 } },
		{
			id: 'low',
			type: 'branch',
			conditions: { type: 'MANAPCT', unit: 'player', operator: 'LESSTHAN', value: 0.2 },
			onTrue: [{ id: 'gem', type: 'item', action: [20520], conditions: [] }],
			onFalse: [
				{
					id: 'fb',
					type: 'spell',
					action: 133,
					ranked: false,
					disabled: true,
					conditions: { type: 'DEBUFF', unit: 'target', spell: 12654, owndebuff: true, operator: 'LESSTHAN', value: 3 },
				},
			],
		},
	],
	cooldowns: [{ id: 'cd', type: 'spell', action: 12472, conditions: { type: 'BUFF', unit: 'player', spell: 11129 } }],
};

describe('import: Rotation Master to an APL', () => {
	const result = importRotationMaster(exportString(BODY), MAGE);
	it('converts the structure', () => {
		expect(result.ok).toBe(true);
		if (!result.ok) return;
		const apl = result.apl;
		expect(apl.type).toBe('TypeAPL');
		// out of combat -> a pre-pull action, at the rank the character casts
		expect(apl.prepullActions).toEqual([{ action: { castSpell: { spellId: { spellId: 25306, rank: 12 } } }, doAtValue: { const: { val: '-1s' } } }]);
		const list = apl.priorityList as Array<Record<string, any>>;
		// cooldowns first, then the rotation; the branch condition rides on both arms
		expect(list.map(i => i.action.castSpell.spellId)).toEqual([{ spellId: 12472 }, { spellId: 11129 }, { itemId: 20520 }, { spellId: 25306, rank: 12 }]);
		expect(list[0].action.condition).toEqual({ auraIsActive: { auraId: { spellId: 11129 } } });
		expect(list[1].action.condition).toEqual({ spellIsReady: { spellId: { spellId: 11129 } } });
		expect(list[2].action.condition).toEqual({ cmp: { op: 'OpLt', lhs: { currentManaPercent: {} }, rhs: { const: { val: '20%' } } } });
		expect(list[3].action.condition.and.vals[0]).toEqual({ not: { val: list[2].action.condition } });
		expect(list[3].action.condition.and.vals[1]).toEqual({ dotIsActive: { spellId: { spellId: 12654 } } });
		// a disabled step is kept, hidden
		expect(list.map(i => i.hide === true)).toEqual([false, false, false, true]);
		expect(result.meta.name).toBe('Fire mage');
	});
	it('says what became of every step', () => {
		if (!result.ok) throw new Error('import failed');
		const byPath = Object.fromEntries(result.items.map(i => [i.path, i]));
		expect(byPath['rotation[2]'].fidelity).toBe('exact');
		expect(byPath['cooldowns[1]'].fidelity).toBe('approximate');
		expect(byPath['rotation[1]'].fidelity).toBe('lossy'); // the seconds before the pull are not kept
		expect(byPath['rotation[1]'].reasons.map(r => r.code)).toContain('x-prepull-time');
		expect(byPath['rotation[3].onFalse[1]'].outcome).toBe('hidden');
		expect(result.summary).toMatchObject({ items: 5, actions: 4, prepull: 1 });
		for (const item of result.items) for (const reason of item.reasons) expect(describeReason(reason).length).toBeGreaterThan(5);
	});
	it('hides the steps that lost something when asked', () => {
		const hidden = importRotationMaster(
			exportString({
				rotation: [
					{ id: 'a', type: 'spell', action: 1, conditions: { type: 'ZONE', value: 'x' } },
					{ id: 'b', type: 'spell', action: 2 },
				],
				cooldowns: [],
			}),
			MAGE,
			{ hideLossy: true },
		);
		if (!hidden.ok) throw new Error('import failed');
		expect(hidden.apl.priorityList.map((i: any) => i.hide === true)).toEqual([true, false]);
		const shown = importRotationMaster(
			exportString({ rotation: [{ id: 'a', type: 'spell', action: 1, conditions: { type: 'ZONE', value: 'x' } }], cooldowns: [] }),
			MAGE,
		);
		expect(shown.ok && shown.apl.priorityList[0].hide).toBeUndefined();
		expect(shown.ok && describeReasons(shown.items[0].reasons)[0]).toContain('Zone');
	});
	it('leaves out what this simulator cannot read, and says so', () => {
		const kinds = { values: new Set(['and', 'or', 'not', 'cmp', 'const', 'auraIsActive']), actions: new Set(['castSpell']) };
		const r = importRotationMaster(exportString(BODY), { ...MAGE, kinds });
		if (!r.ok) throw new Error('import failed');
		const readyStep = r.apl.priorityList.find((i: any) => i.action.castSpell.spellId.spellId === 11129);
		expect(readyStep.action.condition).toBeUndefined();
		expect(r.items.find(i => i.path === 'rotation[2]')?.reasons.map(x => x.code)).toContain('x-cond-dropped');
	});
	it('reads the old export the addon wrote before conditions had ids, and reports every step lossy', () => {
		const r = importRotationMaster(LEGACY_EXPORTS.lzw, { repo: 'classic', edition: 'era', cls: 'SHAMAN', spec: 'SpecEnhancementShaman' });
		if (!r.ok) throw new Error('import failed');
		expect(r.summary.lossy).toBe(12);
		expect(r.apl.priorityList).toHaveLength(12);
	});
	it('refuses what is not for this simulator or not a rotation', () => {
		const envelope = (applicability: RmValue): string => exportString({ format: 'RotationMaster', schemaVersion: 1, applicability, body: BODY });
		const wrongClass = importRotationMaster(envelope({ class: 'WARRIOR', families: ['era'], actionContext: 'player' }), MAGE);
		expect(!wrongClass.ok && wrongClass.error.code).toBe('class-mismatch');
		const wrongEdition = importRotationMaster(envelope({ class: 'MAGE', families: ['forever'], actionContext: 'player' }), MAGE);
		expect(!wrongEdition.ok && wrongEdition.error.code).toBe('edition-mismatch');
		expect(importRotationMaster(envelope({ class: 'MAGE', families: ['era'], actionContext: 'player' }), MAGE).ok).toBe(true);
		const notRotation = importRotationMaster(exportString({ format: 'RotationMaster', schemaVersion: 1, id: 'x' }), MAGE);
		expect(!notRotation.ok && notRotation.error.code).toBe('not-a-rotation');
		const garbage = importRotationMaster('definitely not an export', MAGE);
		expect(!garbage.ok && describeError(garbage.error)).toContain('not a Rotation Master export');
		const empty = importRotationMaster(exportString({ rotation: [{ id: 'a', type: 'pet', action: 5 }], cooldowns: [] }), MAGE);
		expect(!empty.ok && empty.error.code).toBe('nothing-imported');
	});
	it('warns when the rotation is for another specialization, and imports it anyway', () => {
		const other = importRotationMaster(exportString({ ...(BODY as object), switch: { type: 'PRIMARY_TALENT_TREE', tree: 3 } } as RmValue), MAGE);
		expect(other.ok && other.warnings).toEqual([{ code: 'spec-mismatch', params: { rotation: 'frost', simulator: 'fire' } }]);
		const same = importRotationMaster(exportString({ ...(BODY as object), switch: { type: 'PRIMARY_TALENT_TREE', tree: 2 } } as RmValue), MAGE);
		expect(same.ok && same.warnings).toEqual([]);
	});
	it('never lets a hostile id or key reach the result as anything but data', () => {
		const r = importRotationMaster(
			exportString({
				rotation: [
					{
						id: '__proto__',
						type: 'spell',
						action: 133,
						conditions: { type: 'BUFF', unit: 'player', spell: 1, __proto__: { polluted: true } } as any,
					},
				],
				cooldowns: [],
			}),
			MAGE,
		);
		expect(r.ok).toBe(true);
		expect(({} as any).polluted).toBeUndefined();
	});
});

const APL = {
	type: 'TypeAPL',
	prepullActions: [{ action: { castSpell: { spellId: { spellId: 25306, rank: 12 } } }, doAtValue: { const: { val: '-2s' } } }],
	priorityList: [
		{ action: { condition: { spellIsReady: { spellId: { spellId: 11129 } } }, castSpell: { spellId: { spellId: 11129 } } } },
		{ hide: true, action: { castSpell: { spellId: { spellId: 12472 } } } },
		{
			action: {
				condition: { cmp: { op: 'OpGt', lhs: { remainingTime: {} }, rhs: { const: { val: '10s' } } } },
				castSpell: { spellId: { spellId: 12051 } },
			},
		},
		{ action: { condition: { numberTargets: {} }, castSpell: { spellId: { spellId: 10 } } } },
		{ action: { autocastOtherCooldowns: {} } },
	],
};

describe('export: an APL to Rotation Master', () => {
	const result = exportRotationMaster(APL, MAGE);
	it('writes a string the addon reads, scoped to the specialization', () => {
		expect(result.ok).toBe(true);
		if (!result.ok) return;
		const back = decodeRmExport(result.text);
		expect(back.ok && back.value).toEqual(result.body);
		expect(result.body.switch).toEqual({ type: 'PRIMARY_TALENT_TREE', tree: 2 });
		expect(result.body.tags).toEqual(['wowsims']);
		expect(result.body.cooldowns).toEqual([]);
		for (const step of result.body.rotation) expect(typeof step.id).toBe('string');
	});
	it('turns the pre-pull actions into steps that apply out of combat, ahead of the list', () => {
		if (!result.ok) throw new Error('export failed');
		expect(result.body.rotation[0]).toMatchObject({ type: 'spell', action: 25306, conditions: { type: 'NOT', value: { type: 'COMBAT', unit: 'player' } } });
		expect(result.items[0].path).toBe('prepullActions[1]');
		expect(result.items[0].reasons.map(r => r.code)).toContain('prepull-timing');
	});
	it('keeps a hidden item as a disabled step, and disables what is not an exact copy', () => {
		if (!result.ok) throw new Error('export failed');
		const byAction = (id: number) => result.body.rotation.find((s: any) => s.action === id);
		expect(byAction(11129).disabled).toBeUndefined();
		expect(byAction(12472).disabled).toBe(true);
		expect(byAction(12051).disabled).toBe(true); // the fight length has no counterpart
		expect(byAction(10).disabled).toBe(true); // the number of targets is only approximately the enemies you have threat on
		expect(result.items.find(i => i.path === 'priorityList[5]')?.outcome).toBe('skipped');
		expect(result.summary).toMatchObject({ steps: 5, skipped: 1 });
		const loose = exportRotationMaster(APL, MAGE, { enableApproximate: true });
		if (!loose.ok) throw new Error('export failed');
		expect(loose.body.rotation.find((s: any) => s.action === 10).disabled).toBeUndefined();
	});
	it('records spells without a rank on editions that have them', () => {
		if (!result.ok) throw new Error('export failed');
		expect(result.body.rotation.every((s: any) => s.type !== 'spell' || s.ranked === false)).toBe(true);
	});
	it('scopes by spec id on the progression editions and leaves out the scope when the spec is unknown', () => {
		const mop = exportRotationMaster(APL, { repo: 'mop', edition: 'progression', cls: 'MAGE', spec: 'SpecFrostMage' });
		expect(mop.ok && mop.body.specs).toEqual([64]);
		expect(mop.ok && mop.body.switch).toBeUndefined();
		const unknown = exportRotationMaster(APL, { repo: 'mop', edition: 'progression', cls: 'MAGE' });
		expect(unknown.ok && unknown.body.specs).toBeUndefined();
	});
	it('refuses a rotation that is not an APL, or has nothing to export', () => {
		const auto = exportRotationMaster({ type: 'TypeAuto' }, MAGE);
		expect(!auto.ok && auto.error.code).toBe('not-an-apl');
		expect(!auto.ok && describeError(auto.error)).toContain('TypeAuto');
		const none = exportRotationMaster({ type: 'TypeAPL', priorityList: [{ action: { autocastOtherCooldowns: {} } }] }, MAGE);
		expect(!none.ok && none.error.code).toBe('nothing-exported');
		const blank = exportRotationMaster({ type: 'TypeAPL' }, MAGE);
		expect(!blank.ok && describeError(blank.error)).toBe('This APL has no actions.');
	});
});

describe('round trip', () => {
	it('keeps the ids, order and exact conditions of an APL through Rotation Master and back', () => {
		const apl = {
			type: 'TypeAPL',
			priorityList: [
				{
					action: {
						condition: {
							and: {
								vals: [{ spellIsReady: { spellId: { spellId: 11129 } } }, { not: { val: { auraIsActive: { auraId: { spellId: 12472 } } } } }],
							},
						},
						castSpell: { spellId: { spellId: 11129 } },
					},
				},
				{
					action: {
						condition: { cmp: { op: 'OpLt', lhs: { currentManaPercent: {} }, rhs: { const: { val: '20%' } } } },
						castSpell: { spellId: { itemId: 20520 } },
					},
				},
				{ action: { castSpell: { spellId: { spellId: 116 } } } },
			],
		};
		const out = exportRotationMaster(apl, { ...MAGE, ladders: undefined });
		if (!out.ok) throw new Error('export failed');
		const back = importRotationMaster(out.text, { ...MAGE, ladders: undefined });
		if (!back.ok) throw new Error('import failed');
		expect(back.apl.priorityList).toEqual(apl.priorityList);
		expect(back.summary).toMatchObject({ exact: 3, approximate: 0, lossy: 0 });
	});
});

describe('the report rows', () => {
	it('lists the steps that need a look first, with the reasons in words', () => {
		const imported = importRotationMaster(
			exportString({
				rotation: [
					{ id: 'a', type: 'spell', action: 1 },
					{ id: 'b', type: 'spell', action: 2, conditions: { type: 'ZONE', value: 'x' } },
				],
				cooldowns: [],
			}),
			MAGE,
		);
		if (!imported.ok) throw new Error('import failed');
		const rows = importRows(imported.items);
		expect(rows.map(r => r.fidelity)).toEqual(['lossy', 'exact']);
		expect(rows[0].sentences[0]).toContain('Zone');
		expect(rows[1].sentences).toEqual([]);
		const exported = exportRotationMaster(APL, MAGE);
		if (!exported.ok) throw new Error('export failed');
		const out = exportRows(exported.items);
		expect(out[0].fidelity).not.toBe('exact');
		expect(out.find(r => r.sentences.some(s => s.includes('fight length')))).toBeDefined();
	});
});

describe('names and ladders', () => {
	it('finds the specialization a word names', () => {
		expect(specSlug('SHAMAN', 'SpecElementalShaman')).toBe('elemental');
		expect(specSlug('DRUID', 'SpecFeralTankDruid')).toBe('guardian');
		expect(specSlug('HUNTER', 'SpecBeastMasteryHunter')).toBe('beast-mastery');
		expect(specSlug('MAGE', undefined)).toBeNull();
		expect(specSlug('DRUID', 'SpecFeralCatDruid')).toBe('feral');
	});
	it('does not mind a spec that names no specialization', () => {
		expect(specInfo('MAGE', 'SpecMage')).toEqual({ named: false, slug: null });
		expect(specInfo('WARRIOR', 'SpecDpsWarrior')).toEqual({ named: false, slug: null });
		expect(specInfo('MAGE', 'SpecSpellslinger')).toEqual({ named: true, slug: null });
		const out = exportRotationMaster(APL, { repo: 'tbc-new', edition: 'tbc', cls: 'MAGE', spec: 'SpecMage' });
		expect(out.ok && out.warnings).toEqual([]);
		expect(out.ok && out.body.switch).toBeUndefined();
		const odd = exportRotationMaster(APL, { repo: 'tbc-new', edition: 'tbc', cls: 'MAGE', spec: 'SpecSpellslinger' });
		expect(odd.ok && odd.warnings).toEqual([{ code: 'spec-unreadable', params: {} }]);
	});
	it('builds rank ladders from a spell list', () => {
		const ladders = laddersOf([
			{ id: 10, name: 'Blizzard', rank: 1 },
			{ id: 6141, name: 'Blizzard', rank: 2 },
			{ id: 99, name: 'Fireblast', rank: 1 },
			{ id: 5, name: 'Twice', rank: 1 },
			{ id: 6, name: 'Twice', rank: 1 },
		]);
		expect(ladders).toEqual([
			[
				[10, 1],
				[6141, 2],
			],
		]);
		expect(ladderIndex(ladders).get(6141)).toBe(ladders[0]);
	});
});

// Every preset rotation this repository ships goes through the converters without throwing, and what comes out is a
// string the addon's format can carry.
const REPOS: Record<string, { repo: SimRepo; edition: Edition }> = {
	classic: { repo: 'classic', edition: 'era' },
	sod: { repo: 'sod', edition: 'era' },
	// the TBC simulator with an APL is wowsims/tbc-new, whose Go module is still named tbc
	tbc: { repo: 'tbc-new', edition: 'tbc' },
	wotlk: { repo: 'wotlk', edition: 'progression' },
	cata: { repo: 'cata', edition: 'progression' },
	mop: { repo: 'mop', edition: 'progression' },
	forever: { repo: 'forever', edition: 'forever' },
};
const CLASS_WORDS: Array<[string, ClassKey]> = [
	['deathknight', 'DEATHKNIGHT'],
	['death_knight', 'DEATHKNIGHT'],
	['druid', 'DRUID'],
	['hunter', 'HUNTER'],
	['mage', 'MAGE'],
	['monk', 'MONK'],
	['paladin', 'PALADIN'],
	['priest', 'PRIEST'],
	['rogue', 'ROGUE'],
	['shaman', 'SHAMAN'],
	['warlock', 'WARLOCK'],
	['warrior', 'WARRIOR'],
];
const testDir = dirname(fileURLToPath(import.meta.url));
const repoRoot = (() => {
	for (let dir = testDir; dir !== dirname(dir); dir = dirname(dir)) if (existsSync(join(dir, 'ui')) && existsSync(join(dir, 'proto'))) return dir;
	return null;
})();
const repoOf = (root: string): { repo: SimRepo; edition: Edition } | null => {
	try {
		const module = /^module\s+\S+\/([\w-]+)\s*$/m.exec(readFileSync(join(root, 'go.mod'), 'utf8'))?.[1];
		return (module && REPOS[module]) || null;
	} catch {
		return null;
	}
};
const presetFiles = (dir: string): string[] =>
	readdirSync(dir).flatMap(entry => {
		const path = join(dir, entry);
		if (statSync(path).isDirectory()) return entry === 'node_modules' ? [] : presetFiles(path);
		return entry.endsWith('.apl.json') ? [path] : [];
	});
const classOfPath = (path: string): ClassKey | null => {
	const parts = path.toLowerCase().split('/');
	return CLASS_WORDS.find(([word]) => parts.some(p => p === word || p.split('_').includes(word) || p.replace(/_/g, '') === word))?.[1] ?? null;
};
const expectSentences = (items: ReadonlyArray<{ reasons: ReadonlyArray<Reason> }>): void => {
	for (const item of items) {
		for (const reason of item.reasons) {
			if (reason.level !== 'check') expect(describeReason(reason), reason.code).not.toBe(reason.code);
		}
	}
};
const here = repoRoot ? repoOf(repoRoot) : null;
describe.skipIf(!here)("this repository's preset rotations", () => {
	const files = repoRoot ? presetFiles(join(repoRoot, 'ui')) : [];
	it('has presets', () => expect(files.length).toBeGreaterThan(0));
	for (const file of files) {
		it(file.slice((repoRoot as string).length + 1), () => {
			const { repo, edition } = here as { repo: SimRepo; edition: Edition };
			const cls = classOfPath(file.slice((repoRoot as string).length + 1));
			if (!cls) return;
			const apl = JSON.parse(readFileSync(file, 'utf8'));
			const converted = aplToRotation(apl, { edition, cls });
			expect(converted.summary.steps).toBe(converted.steps.length);
			// every reason the converter gives has a sentence of its own
			expectSentences(converted.items);
			const out = exportRotationMaster(apl, { repo, edition, cls });
			if (out.ok) {
				expect(decodeRmExport(out.text).ok).toBe(true);
				const again = importRotationMaster(out.text, { repo, edition, cls });
				expect(again.ok).toBe(true);
				if (again.ok) expectSentences(again.items);
				if (again.ok) expect(again.summary.actions + again.summary.prepull).toBeGreaterThan(0);
			} else expect(['nothing-exported', 'not-an-apl', 'empty-apl']).toContain(out.error.code);
		});
	}
});
