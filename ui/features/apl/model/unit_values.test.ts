import { UnitReference, UnitReference_Type as UnitType } from '@generated/proto/common';
import { ActionId, defaultTargetIcon, getPetIconFromName } from '@sim/proto/action_id';
import { makePlayer } from '@testing/apl/player';
import { describe, expect, it, vi } from 'vitest';

import { refToValue, unitOptionModels, withIconUrl } from './unit_values';

vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));

const label = (key: string) => `rotation_tab.apl.helpers.unit_labels.${key}`;

describe('refToValue', () => {
	it.each([
		['an undefined ref', undefined],
		['an Unknown-type ref', UnitReference.create({ type: UnitType.Unknown })],
	])('defaults %s to self or current target depending on targetUI', (_case, ref) => {
		const player: any = makePlayer({} as any);
		expect(refToValue(ref, player, false)).toEqual({ value: ref, icon: 'user', text: label('self') });
		expect(refToValue(ref, player, true)).toEqual({ value: ref, icon: 'bullseye', text: label('current_target') });
	});

	it('resolves a Player reference through the raid, and falls back to a bare value when the index is empty', () => {
		const player: any = makePlayer({} as any);
		const other = { getSpecIcon: () => 'icon-url' };
		player.sim.raid.getPlayer = (index: number) => (index === 0 ? other : undefined);

		const found = UnitReference.create({ type: UnitType.Player, index: 0 });
		expect(refToValue(found, player, false)).toEqual({ value: found, iconUrl: 'icon-url', text: `${label('player')} 1` });

		const missing = UnitReference.create({ type: UnitType.Player, index: 5 });
		expect(refToValue(missing, player, false)).toEqual({ value: missing });
	});

	it('resolves a Target reference through the encounter, and falls back to a bare value when the index is empty', () => {
		const player: any = makePlayer({} as any);
		player.sim.encounter.targetsMetadata.asList = () => [{}];

		const found = UnitReference.create({ type: UnitType.Target, index: 0 });
		expect(refToValue(found, player, false)).toEqual({ value: found, iconUrl: defaultTargetIcon, text: `${label('target')} 1` });

		const missing = UnitReference.create({ type: UnitType.Target, index: 3 });
		expect(refToValue(missing, player, false)).toEqual({ value: missing });
	});

	it('takes a pet reference’s name from after the " - " separator in its metadata', () => {
		const player: any = makePlayer({} as any);
		player.sim.getUnitMetadata = () => ({ getName: () => 'Player 1 - Wolf' });

		const ref = UnitReference.create({ type: UnitType.Pet, index: 0 });
		const result = refToValue(ref, player, false);
		expect(result.text).toBe('Wolf');
		expect(result.iconUrl).toBe(getPetIconFromName('Wolf'));
	});

	it('falls back to the paw icon and an indexed label when the pet has no metadata name', () => {
		const player: any = makePlayer({} as any);
		player.sim.getUnitMetadata = () => ({ getName: () => undefined });

		const ref = UnitReference.create({ type: UnitType.Pet, index: 2 });
		expect(refToValue(ref, player, false)).toEqual({ value: ref, icon: 'paw', text: `${label('pet')} 3` });
	});
});

describe('withIconUrl', () => {
	it('keeps a url or glyph icon, and swaps an ActionId icon for the url it resolved to', () => {
		const ref = UnitReference.create({ type: UnitType.Pet, index: 0 });
		const pet = { value: ref, text: 'Stampede', iconUrl: ActionId.fromSpellId(121818) };

		expect(withIconUrl({ value: ref, icon: 'paw' })).toEqual({ value: ref, icon: 'paw' });
		expect(withIconUrl(pet, { iconUrl: 'stampede.jpg', name: 'Stampede' })).toEqual({ value: ref, text: 'Stampede', iconUrl: 'stampede.jpg' });
		expect(withIconUrl(pet)).toEqual({ value: ref, text: 'Stampede', iconUrl: undefined });
	});
});

describe('unitOptionModels', () => {
	it.each([
		['under its owner when the set uses target UI', 'aura_sources_targets_first', UnitReference.create({ type: UnitType.Self }), true],
		['under the default unit when the set does not use target UI', 'aura_sources', undefined, undefined],
	] as const)('files a pet %s', (_case, unitSet, ownerRef, targetUI) => {
		const player: any = makePlayer({} as any);
		player.getPetMetadatas = () => ({ asList: () => [{}] });
		player.sim.raid.getActivePlayers = () => [];
		player.sim.encounter.targetsMetadata.asList = () => [];
		player.sim.getUnitMetadata = () => ({ getName: () => 'Player 1 - Wolf' });

		const models = unitOptionModels(unitSet, player);
		const petModel = models.find(model => model.unit.text === 'Wolf');
		expect(petModel?.submenu).toEqual([refToValue(ownerRef, player, targetUI)]);
	});

	it('offers self and own pets only for self_and_pets', () => {
		const player: any = makePlayer({} as any);
		player.getPetMetadatas = () => ({ asList: () => [{}, {}] });
		player.sim.raid.getActivePlayers = () => [{ getRaidIndex: () => 1 }];
		player.sim.encounter.targetsMetadata.asList = () => [{}];
		player.sim.getUnitMetadata = () => undefined;

		const self = UnitReference.create({ type: UnitType.Self });
		expect(unitOptionModels('self_and_pets', player).map(model => model.unit.value)).toEqual([
			undefined,
			UnitReference.create({ type: UnitType.Pet, index: 0, owner: self }),
			UnitReference.create({ type: UnitType.Pet, index: 1, owner: self }),
		]);
	});

	it('gives a non-pet unit no submenu', () => {
		const player: any = makePlayer({} as any);
		player.getPetMetadatas = () => ({ asList: () => [] });
		player.sim.raid.getActivePlayers = () => [];
		player.sim.encounter.targetsMetadata.asList = () => [];

		const models = unitOptionModels('aura_sources', player);
		expect(models.every(model => model.submenu === undefined)).toBe(true);
	});
});
