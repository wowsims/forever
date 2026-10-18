import { APLListItem, APLRotation, APLRotation_Type } from '@generated/proto/apl';
import { ActionID, Cooldowns, type Spec } from '@generated/proto/common';
import { describe, expect, it } from 'vitest';

import type { Player } from '../player/player';
import { hideScheduledCooldownCasts, isEqualAPLRotation, renameAPLReference } from './apl_utils';

// Only the TypeSimple branch reaches the player, and none of these cases take it.
const player = null as unknown as Player<Spec>;

const rotation = (type: APLRotation_Type, spellId: number) =>
	APLRotation.create({
		type,
		priorityList: [
			{ action: { action: { oneofKind: 'castSpell', castSpell: { spellId: ActionID.create({ rawId: { oneofKind: 'spellId', spellId } }) } } } },
		],
	});

describe('isEqualAPLRotation', () => {
	it('matches an Auto rotation against the APL rotation it resolves to', () => {
		expect(isEqualAPLRotation(player, rotation(APLRotation_Type.TypeAuto, 100), rotation(APLRotation_Type.TypeAPL, 100))).toBe(true);
		expect(isEqualAPLRotation(player, rotation(APLRotation_Type.TypeAPL, 100), rotation(APLRotation_Type.TypeAuto, 100))).toBe(true);
	});

	it('still separates rotations that differ below the type', () => {
		expect(isEqualAPLRotation(player, rotation(APLRotation_Type.TypeAuto, 100), rotation(APLRotation_Type.TypeAPL, 200))).toBe(false);
	});

	it('leaves both arguments untouched', () => {
		const auto = rotation(APLRotation_Type.TypeAuto, 100);
		const apl = rotation(APLRotation_Type.TypeAPL, 100);
		isEqualAPLRotation(player, auto, apl);
		expect(auto.type).toBe(APLRotation_Type.TypeAuto);
		expect(apl.type).toBe(APLRotation_Type.TypeAPL);
	});

	it('treats a missing rotation as unequal to a present one', () => {
		expect(isEqualAPLRotation(player, undefined, rotation(APLRotation_Type.TypeAPL, 100))).toBe(false);
		expect(isEqualAPLRotation(player, undefined, undefined)).toBe(true);
	});

	// The Simple branch does reach the player, so it needs the two codec hooks it calls.
	const simplePlayer = {
		specTypeFunctions: { rotationFromJson: (json: unknown) => json, rotationEquals: (a: unknown, b: unknown) => JSON.stringify(a) === JSON.stringify(b) },
	} as unknown as Player<Spec>;

	const simple = (cooldowns: Cooldowns) =>
		APLRotation.create({ type: APLRotation_Type.TypeSimple, simple: { specRotationJson: '{"spec":"Fury"}', cooldowns } });

	it('separates Simple rotations that differ only in cooldown timings', () => {
		const a = simple(Cooldowns.create({ cooldowns: [{ id: ActionID.create(), timings: [10] }] }));
		const b = simple(Cooldowns.create({ cooldowns: [{ id: ActionID.create(), timings: [20] }] }));

		expect(isEqualAPLRotation(simplePlayer, a, b)).toBe(false);
		expect(isEqualAPLRotation(simplePlayer, a, simple(Cooldowns.clone(a.simple!.cooldowns!)))).toBe(true);
	});

	// A Simple rotation only gets its spec json once a checkbox is touched, so picking a cooldown first leaves it empty.
	it('reads an empty spec json as the default spec rotation', () => {
		const defaultsPlayer = {
			specTypeFunctions: { ...simplePlayer.specTypeFunctions, rotationCreate: () => ({}) },
		} as unknown as Player<Spec>;
		const picked = Cooldowns.create({ cooldowns: [{ id: ActionID.create({ rawId: { oneofKind: 'spellId', spellId: 13877 } }) }] });
		const current = APLRotation.create({ type: APLRotation_Type.TypeSimple, simple: { specRotationJson: '', cooldowns: picked } });
		const preset = (specRotationJson: string, cooldowns: Cooldowns) =>
			APLRotation.create({ type: APLRotation_Type.TypeSimple, simple: { specRotationJson, cooldowns } });

		expect(isEqualAPLRotation(defaultsPlayer, current, preset('{"spec":"Fury"}', Cooldowns.clone(picked)))).toBe(false);
		expect(isEqualAPLRotation(defaultsPlayer, current, preset('{}', Cooldowns.clone(picked)))).toBe(true);
		expect(isEqualAPLRotation(defaultsPlayer, preset('{}', Cooldowns.clone(picked)), current)).toBe(true);
	});
});

describe('renameAPLReference', () => {
	// `Action Group is used` stores the group as a bare `{name}`, which the generic walk cannot spot.
	it('renames a group inside an actionGroupUsed condition', () => {
		const rotation = APLRotation.create({
			priorityList: [{ action: { condition: { value: { oneofKind: 'actionGroupUsed', actionGroupUsed: { name: 'Cooldowns' } } } } }],
		});

		renameAPLReference(rotation, { type: 'group', oldName: 'Cooldowns', newName: 'Burst' });

		const value = rotation.priorityList[0].action!.condition!.value;
		expect(value.oneofKind === 'actionGroupUsed' && value.actionGroupUsed.name).toBe('Burst');
	});
});

describe('hideScheduledCooldownCasts', () => {
	const ICY_VEINS = 12472;
	const COLD_SNAP = 11958;
	const BERSERKING = 20554;
	const ARCANE_POWER = 12042;
	const MANA_EMERALD = 22044;

	const spell = (spellId: number, tag = 0) => ActionID.create({ rawId: { oneofKind: 'spellId', spellId }, tag });
	const item = (itemId: number) => ActionID.create({ rawId: { oneofKind: 'itemId', itemId } });
	const cast = (spellId: ActionID) => APLListItem.create({ action: { action: { oneofKind: 'castSpell', castSpell: { spellId } } } });
	// Cold Snap gated on Icy Veins being on cooldown: it reads Icy Veins but does not cast it.
	const coldSnap = () =>
		APLListItem.create({
			action: {
				condition: { value: { oneofKind: 'spellIsReady', spellIsReady: { spellId: spell(ICY_VEINS) } } },
				action: { oneofKind: 'castSpell', castSpell: { spellId: spell(COLD_SNAP) } },
			},
		});
	const preset = () =>
		APLRotation.create({
			priorityList: [coldSnap(), cast(spell(ICY_VEINS)), cast(spell(BERSERKING, 1)), cast(item(MANA_EMERALD))],
			groups: [{ name: 'Burst', actions: [cast(spell(ICY_VEINS)), cast(spell(COLD_SNAP))] }],
		});

	const cd = (id: ActionID, timings: Array<number>) => ({ id, timings });
	const hide = (cooldowns: Cooldowns) => {
		const rotation = preset();
		hideScheduledCooldownCasts(rotation, cooldowns);
		return rotation;
	};
	const hidden = (rotation: APLRotation) => [rotation.priorityList.map(i => i.hide), rotation.groups[0].actions.map(i => i.hide)];

	it('leaves the preset untouched when no cooldown has timings', () => {
		expect(hide(Cooldowns.create())).toEqual(preset());
		// A picked cooldown with no timings yields an invalid schedule, so the preset's line must stay.
		expect(hide(Cooldowns.create({ cooldowns: [cd(spell(ICY_VEINS), [])] }))).toEqual(preset());
	});

	it('hides every Icy Veins cast in the priority list and in groups once Icy Veins is scheduled, and nothing else', () => {
		const rotation = hide(Cooldowns.create({ cooldowns: [cd(spell(ICY_VEINS), [10])] }));

		// Cold Snap only reads Icy Veins in its condition, so it stays to enable a second scheduled Icy Veins.
		expect(hidden(rotation)).toEqual([
			[false, true, false, false],
			[true, false],
		]);
		// Nothing is removed, so indices stay stable.
		expect(rotation.priorityList).toHaveLength(preset().priorityList.length);
	});

	it('matches across tags so a differently tagged variant still replaces the preset line', () => {
		expect(hidden(hide(Cooldowns.create({ cooldowns: [cd(spell(BERSERKING, 0), [5])] })))[0]).toEqual([false, false, true, false]);
	});

	it('hides item casts by item id', () => {
		expect(hidden(hide(Cooldowns.create({ cooldowns: [cd(item(MANA_EMERALD), [30])] })))[0]).toEqual([false, false, false, true]);
	});

	it('is a no-op for a scheduled cooldown the preset never casts directly', () => {
		expect(hide(Cooldowns.create({ cooldowns: [cd(spell(ARCANE_POWER), [15])] }))).toEqual(preset());
	});
});
