// What this pins is the promise of the Rotation Master import: it replaces the APL and touches nothing else of the
// player. The player below throws on any member the import has no business reading, so a stray write fails loudly.
import { APLRotation, APLRotation_Type } from '@generated/proto/apl';
import { Class, Spec } from '@generated/proto/common';
import { fakeHost } from '@testing/sim/fake_host';
import { describe, expect, it, vi } from 'vitest';

import { decodeRmExport, encodeRmExport, type RmValue } from '../rotation_master';
import { applyRotationMasterImport, checkRotationMasterImport, exportRotationMasterText, rotationMasterContext } from './rotation_master';

// The real one gates store notifications; here it only has to run the body.
vi.mock('@sim/state/batch', () => ({ batch: (body: () => void) => body() }));

const strict = <T extends object>(name: string, members: T): T =>
	new Proxy(members, {
		get: (target, key) => {
			if (key in target) return (target as Record<PropertyKey, unknown>)[key];
			if (key === 'then' || typeof key === 'symbol') return undefined;
			throw new Error(`the Rotation Master import touched ${name}.${String(key)}`);
		},
	});

const SIMPLE = '{"conserveStart":20}';
const SPELLS = [
	{ id: 133, name: 'Fireball', rank: 1 },
	{ id: 143, name: 'Fireball', rank: 2 },
	{ id: 25306, name: 'Fireball', rank: 12 },
];
const fakePlayer = (rotation: APLRotation, cls = Class.ClassMage, spec = Spec.SpecMage) => {
	const state = { rotation, sets: 0 };
	const player = strict('player', {
		getClass: () => cls,
		getSpec: () => spec,
		get aplRotation() {
			return state.rotation;
		},
		getRotationType: () => (state.rotation.type == APLRotation_Type.TypeUnknown ? APLRotation_Type.TypeAPL : state.rotation.type),
		getResolvedAplRotation: () => state.rotation,
		setAplRotation: (next: APLRotation) => {
			state.sets++;
			state.rotation = next;
		},
	});
	return { state, host: fakeHost({ player: player as never, sim: strict('sim', { db: { getSpellIcons: () => SPELLS } }) as never }) };
};

const exportString = (value: RmValue): string => {
	const out = encodeRmExport(value);
	if (!out.ok) throw new Error(out.reason);
	return out.text;
};
const BODY: RmValue = {
	rotation: [
		{ id: 'a', type: 'spell', action: 11129, conditions: { type: 'SPELL_AVAIL', spell: 11129 } },
		{ id: 'b', type: 'spell', action: 133, ranked: false, conditions: { type: 'NOT', value: { type: 'COMBAT', unit: 'player' } } },
	],
	cooldowns: [],
};

describe('rotationMasterContext', () => {
	it('names the simulator, the class, the spec and what the APL can say', () => {
		const { host } = fakePlayer(APLRotation.create());
		const ctx = rotationMasterContext(host);
		expect(ctx).toMatchObject({ repo: 'forever', edition: 'forever', cls: 'MAGE', spec: 'SpecMage' });
		expect(ctx.kinds?.values.has('spellIsReady')).toBe(true);
		expect(ctx.kinds?.values.has('definitelyNotAValue')).toBe(false);
		expect(ctx.kinds?.actions.has('castSpell')).toBe(true);
		// spells have ranks here: a spell is named by its first rank and the simulator needs the highest
		expect(ctx.ladders?.get(133)).toEqual([
			[133, 1],
			[143, 2],
			[25306, 12],
		]);
	});
});

describe('importing a Rotation Master rotation', () => {
	const existing = APLRotation.create({
		type: APLRotation_Type.TypeSimple,
		simple: { specRotationJson: SIMPLE, cooldowns: undefined },
		priorityList: [],
		groups: [{ name: 'old', actions: [], variables: [] } as never],
	});

	it('replaces the APL and only the APL', () => {
		const { host, state } = fakePlayer(existing);
		const checked = checkRotationMasterImport(host, exportString(BODY));
		expect(checked.ok).toBe(true);
		if (!checked.ok) return;
		// checking changes nothing
		expect(state.sets).toBe(0);

		applyRotationMasterImport(host, checked.apl);

		expect(state.sets).toBe(1);
		expect(state.rotation.type).toBe(APLRotation_Type.TypeAPL);
		expect(state.rotation.priorityList).toHaveLength(1);
		expect(state.rotation.prepullActions).toHaveLength(1);
		expect(state.rotation.prepullActions[0].action?.action).toMatchObject({
			oneofKind: 'castSpell',
			castSpell: { spellId: { rawId: { oneofKind: 'spellId', spellId: 25306 } } },
		});
		expect(state.rotation.groups).toEqual([]);
		// the Simple rotation's settings were not the import's to change
		expect(state.rotation.simple?.specRotationJson).toBe(SIMPLE);
	});

	it('never reaches for gear, talents, buffs, the encounter or anything else (the strict player and sim would throw)', () => {
		const { host } = fakePlayer(existing);
		const checked = checkRotationMasterImport(host, exportString(BODY));
		if (!checked.ok) throw new Error('import failed');
		expect(() => applyRotationMasterImport(host, checked.apl)).not.toThrow();
	});

	it('does not apply what it refused', () => {
		const { host, state } = fakePlayer(existing);
		const checked = checkRotationMasterImport(host, 'not an export');
		expect(checked.ok).toBe(false);
		expect(state.sets).toBe(0);
		expect(state.rotation).toBe(existing);
	});

	it('refuses a rotation for another class', () => {
		const { host } = fakePlayer(existing, Class.ClassWarrior, Spec.SpecDpsWarrior);
		const text = exportString({
			format: 'RotationMaster',
			schemaVersion: 1,
			applicability: { class: 'MAGE', families: ['forever'], actionContext: 'player' },
			body: BODY,
		});
		const checked = checkRotationMasterImport(host, text);
		expect(!checked.ok && checked.error.code).toBe('class-mismatch');
	});
});

describe('exporting the rotation', () => {
	it('writes the APL of the player as a string the importer reads back', () => {
		const apl = APLRotation.fromJson({
			type: 'TypeAPL',
			priorityList: [{ action: { condition: { spellIsReady: { spellId: { spellId: 11129 } } }, castSpell: { spellId: { spellId: 11129 } } } }],
		});
		const { host, state } = fakePlayer(apl);
		const made = exportRotationMasterText(host);
		expect(made.generatedFrom).toBeNull();
		expect(made.result.ok).toBe(true);
		if (!made.result.ok) return;
		expect(state.sets).toBe(0);
		expect(decodeRmExport(made.result.text)).toMatchObject({ ok: true, value: { tags: ['wowsims'] } });
		const again = checkRotationMasterImport(host, made.result.text);
		expect(again.ok && again.apl.priorityList).toEqual(JSON.parse(JSON.stringify(APLRotation.toJson(apl))).priorityList);
	});

	it('exports the priority list the simulator generates for an Auto rotation, and says so', () => {
		const generated = APLRotation.fromJson({
			type: 'TypeAuto',
			priorityList: [{ action: { castSpell: { spellId: { spellId: 11129 } } } }],
		});
		const { host } = fakePlayer(generated);
		const made = exportRotationMasterText(host);
		expect(made.generatedFrom).toBe('TypeAuto');
		expect(made.result.ok).toBe(true);
	});
});
