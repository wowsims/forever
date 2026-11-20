import { TristateEffect } from '@generated/proto/common';
import { IndividualSimSettings } from '@generated/proto/ui';
import { describe, expect, it } from 'vitest';

import { fromJsonLenient, fromJsonStringLenient } from './lenient_json';

// A saved rotation casting an APL placeholder the protos do not list, beside one they do.
const savedSettings = JSON.stringify({
	player: {
		rotation: {
			priorityList: [
				{ action: { castSpell: { spellId: { otherId: 'OtherActionRetiredPlaceholder' } } } },
				{ action: { castSpell: { spellId: { spellId: 15208 } } } },
			],
		},
	},
});

describe('fromJsonStringLenient', () => {
	it('drops an unknown enum name in a oneof, so the settings still write to binary', () => {
		const settings = fromJsonStringLenient(IndividualSimSettings, savedSettings);

		expect(() => IndividualSimSettings.toBinary(settings)).not.toThrow();
		const [retired, kept] = settings.player!.rotation!.priorityList;
		expect(retired.action?.action.oneofKind === 'castSpell' && retired.action.action.castSpell.spellId?.rawId).toEqual({ oneofKind: undefined });
		expect(kept.action?.action.oneofKind === 'castSpell' && kept.action.action.castSpell.spellId?.rawId).toEqual({ oneofKind: 'spellId', spellId: 15208 });
	});

	it('reads the same JSON as protobuf-ts does when every name is known', () => {
		const known = savedSettings.replace('OtherActionRetiredPlaceholder', 'OtherActionOffensivePotion');

		expect(fromJsonStringLenient(IndividualSimSettings, known)).toEqual(IndividualSimSettings.fromJsonString(known));
	});

	// A save can hold Judgement of the Crusader as a bool, the type the field has without its improved state.
	it('reads a bool saved in an enum field as its number', () => {
		const saved = { debuffs: { judgementOfTheCrusader: true, faerieFire: true }, player: { name: 'Saved' } };
		const json = JSON.stringify(saved);

		const settings = fromJsonStringLenient(IndividualSimSettings, json);

		expect(settings.debuffs?.judgementOfTheCrusader).toBe(TristateEffect.TristateEffectRegular);
		expect(settings.debuffs?.faerieFire).toBe(true);
		expect(settings.player?.name).toBe('Saved');
		expect(
			fromJsonStringLenient(IndividualSimSettings, json.replace('"judgementOfTheCrusader":true', '"judgementOfTheCrusader":false')).debuffs
				?.judgementOfTheCrusader,
		).toBe(TristateEffect.TristateEffectMissing);
	});

	it('leaves the JSON it was handed untouched', () => {
		const saved = { debuffs: { judgementOfTheCrusader: true } };

		fromJsonLenient(IndividualSimSettings, saved);

		expect(saved.debuffs.judgementOfTheCrusader).toBe(true);
	});
});
