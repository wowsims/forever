import { APLRotation_Type } from '@generated/proto/apl';
import { Debuffs, IndividualBuffs, PartyBuffs } from '@generated/proto/buffs';
import { ConsumesSpec, EquipmentSpec, Profession, Race, Spec, UnitReference, UnitReference_Type } from '@generated/proto/common';
import { Hunter_Rotation_Playstyle } from '@generated/proto/hunter';
import { describe, expect, it } from 'vitest';

import { makePresetBuildFromJSON, specDefaultsFromSettings } from './preset_utils';

// A settings: section sets the spec's options, so these hold them whenever they state the player's settings.
const buildFrom = (player: Record<string, unknown>) =>
	makePresetBuildFromJSON('p', Spec.SpecProtectionWarrior, { player: { class: 'ClassWarrior', protectionWarrior: { options: {} }, ...player } });

describe('makePresetBuildFromJSON player timings', () => {
	// proto3 hands back 0 for every scalar a file does not mention, and copying that 0 into the build would make it
	// indistinguishable from a preset that means 0.
	it('omits the timings a build never states', () => {
		const build = buildFrom({ healingModel: { burstWindow: 6 }, inFrontOfTarget: true });

		expect(build.settings?.playerOptions).not.toHaveProperty('reactionTimeMs');
		expect(build.settings?.playerOptions).not.toHaveProperty('channelClipDelayMs');
	});

	it('keeps the ones a build does state', () => {
		const build = buildFrom({ reactionTimeMs: 100, channelClipDelayMs: 50 });

		expect(build.settings?.playerOptions?.reactionTimeMs).toBe(100);
		expect(build.settings?.playerOptions?.channelClipDelayMs).toBe(50);
	});
});

describe('makePresetBuildFromJSON settings', () => {
	it('states none of the player settings for a file without a settings: section', () => {
		const build = makePresetBuildFromJSON('p', Spec.SpecDpsWarrior, {
			player: { class: 'ClassWarrior', dpsWarrior: {}, talentsString: '1-2-3', consumables: { offensivePotId: 22838 } },
		});

		expect(build.settings).toEqual({ name: 'p', consumables: ConsumesSpec.create({ offensivePotId: 22838 }) });
	});

	it('states the item swap set a file holds', () => {
		const build = buildFrom({ enableItemSwap: true, itemSwap: { items: [{ id: 1 }] } });

		expect(build.settings?.playerOptions?.enableItemSwap).toBe(true);
		expect(build.settings?.playerOptions?.itemSwap?.items).toEqual([expect.objectContaining({ id: 1 })]);
	});

	it('turns item swap off for settings that hold no item swap set', () => {
		expect(buildFrom({}).settings?.playerOptions?.enableItemSwap).toBe(false);
	});
});

describe('makePresetBuildFromJSON rotation', () => {
	it('switches the rotation to Auto for a file that says apl: type auto', () => {
		const build = buildFrom({ rotation: { type: 'TypeAuto' } });

		expect(build.rotationType).toBe(APLRotation_Type.TypeAuto);
		expect(build).not.toHaveProperty('rotation');
	});

	it('keeps the rotation of any other type and leaves the rotation type to it', () => {
		const build = buildFrom({ rotation: { type: 'TypeAPL', priorityList: [] } });

		expect(build.rotation?.rotation.rotation?.type).toBe(APLRotation_Type.TypeAPL);
		expect(build).not.toHaveProperty('rotationType');
	});
});

describe('specDefaultsFromSettings', () => {
	const hunter = {
		player: {
			class: 'ClassHunter',
			race: 'RaceOrc',
			profession1: 'Engineering',
			distanceFromTarget: 11,
			hunter: { options: { classOptions: { petUptime: 1 } } },
			equipment: { items: [{ id: 28275, enchant: 2724 }] },
			talentsString: '-005-55',
			rotation: { type: 'TypeSimple', simple: { specRotationJson: '{"timeToWeave":400,"playstyle":"Weave"}' } },
		},
		raidBuffs: { giftOfTheWild: 1 },
		settings: { iterations: 25000 },
	};

	const warrior = {
		player: { class: 'ClassWarrior', inFrontOfTarget: true, protectionWarrior: {} },
		tanks: [{ type: 'Player' }],
	};

	it('reads the proto-backed defaults a file holds', () => {
		const defaults = specDefaultsFromSettings(Spec.SpecHunter, hunter);

		expect(defaults.gear).toEqual(EquipmentSpec.fromJson({ items: [{ id: 28275, enchant: 2724 }] }));
		expect(defaults.talents.talentsString).toBe('-005-55');
		expect(defaults.specOptions.classOptions?.petUptime).toBe(1);
		expect(defaults.raidBuffs.giftOfTheWild).toBe(1);
		expect(defaults.rotationType).toBe(APLRotation_Type.TypeSimple);
		expect(defaults.simpleRotation).toMatchObject({ timeToWeave: 400, playstyle: Hunter_Rotation_Playstyle.Weave });
		expect(defaults).not.toHaveProperty('aplRotation');
		expect(defaults.other).toMatchObject({ race: Race.RaceOrc, profession1: Profession.Engineering, distanceFromTarget: 11, iterationCount: 25000 });
	});

	it('gives an empty message for each one the file does not hold', () => {
		const defaults = specDefaultsFromSettings(Spec.SpecHunter, { player: { class: 'ClassHunter', hunter: {} } });

		expect(defaults.consumables).toEqual(ConsumesSpec.create());
		expect(defaults.other).toMatchObject({ inFrontOfTarget: false, tanks: [] });
		expect(defaults.debuffs).toEqual(Debuffs.create());
		expect(defaults.partyBuffs).toEqual(PartyBuffs.create());
		expect(defaults.individualBuffs).toEqual(IndividualBuffs.create());
		expect(defaults).not.toHaveProperty('rotationType');
	});

	it('reads where the player stands and whom the raid tanks with', () => {
		const defaults = specDefaultsFromSettings(Spec.SpecProtectionWarrior, warrior);

		expect(defaults.other).toMatchObject({ inFrontOfTarget: true, tanks: [UnitReference.create({ type: UnitReference_Type.Player })] });
	});

	it('keeps the priority list of an APL rotation', () => {
		const defaults = specDefaultsFromSettings(Spec.SpecHunter, {
			player: { hunter: {}, rotation: { type: 'TypeAPL', priorityList: [{ action: { castSpell: { spellId: { spellId: 75 } } } }] } },
		});

		expect(defaults.rotationType).toBe(APLRotation_Type.TypeAPL);
		expect(defaults.aplRotation?.priorityList).toHaveLength(1);
	});

	it('refuses the settings of another spec', () => {
		expect(() => specDefaultsFromSettings(Spec.SpecHunter, { player: { dpsWarrior: {} } })).toThrow('not SpecHunter settings');
	});
});
