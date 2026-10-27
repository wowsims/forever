import { ReforgeSettings as ReforgeSettingsProto } from '@generated/proto/api';
import { APLRotation, APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import { ConsumesSpec, ItemSlot } from '@generated/proto/common';
import type { PresetBuild } from '@sim/presets/types';
import { DENSE_SHARPENING_STONE_ID } from '@sim/proto/sim_tables_auto_gen';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { swapSet, SWORD, weaponsDatabase } from '@testing/sim/weapons';
import { beforeAll, describe, expect, it, vi } from 'vitest';

import { applyBuild } from './apply_build';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

beforeAll(() => serveDatabase(weaponsDatabase()));

const setAplRotation = vi.fn();
const modifyAplRotation = vi.fn((modify: (rotation: APLRotation) => void) => modify(live));
const applyPresetSettings = vi.fn();
const fromProto = vi.fn();
const setReactionTime = vi.fn();
const setChannelClipDelay = vi.fn();

let live: APLRotation;

const host = () => {
	live = APLRotation.create({ type: APLRotationType.TypeAPL });
	setAplRotation.mockClear();
	modifyAplRotation.mockClear();
	applyPresetSettings.mockClear();
	fromProto.mockClear();
	setReactionTime.mockClear();
	setChannelClipDelay.mockClear();
	return {
		player: {
			setAplRotation,
			modifyAplRotation,
			setReactionTime,
			setChannelClipDelay,
			setRace: vi.fn(),
			itemSwapSettings: { setEnableItemSwap: vi.fn() },
		},
		sim: {},
		reforger: { applyPresetSettings, fromProto },
	} as any;
};

// What `makePresetSimpleRotation` produces, as a preset build carries it.
const simpleRotation = (json: string) =>
	({
		name: 'p',
		rotation: {
			rotation: APLRotation.create({ type: APLRotationType.TypeSimple, simple: { specRotationJson: json } }),
		},
	}) as unknown as NonNullable<PresetBuild['rotation']>;

describe('applyBuild rotation', () => {
	// Builds are hand-written and set both. Taking only the type branch would leave the
	// previous spec rotation in place — apply "Arms", keep simming Fury.
	it('applies the rotation when the build carries both a type and a rotation', () => {
		applyBuild({ name: 'Arms', rotationType: APLRotationType.TypeSimple, rotation: simpleRotation('{"spec":"Arms"}') }, host());

		expect(setAplRotation).toHaveBeenCalledTimes(1);
		expect(setAplRotation.mock.calls[0][0].simple?.specRotationJson).toBe('{"spec":"Arms"}');
	});

	it('sets the type in place when the build carries only a type', () => {
		applyBuild({ name: 'Auto', rotationType: APLRotationType.TypeAuto }, host());

		expect(setAplRotation).not.toHaveBeenCalled();
		expect(live.type).toBe(APLRotationType.TypeAuto);
	});

	it('applies the rotation alone when the build carries no type', () => {
		applyBuild({ name: 'Simple', rotation: simpleRotation('{"spec":"Fury"}') }, host());

		expect(setAplRotation).toHaveBeenCalledTimes(1);
		expect(modifyAplRotation).not.toHaveBeenCalled();
	});
});

describe('applyBuild player timings', () => {
	it('leaves them alone when the preset states neither', () => {
		applyBuild({ name: 'Magtheridon', settings: { name: 'm', playerOptions: { inFrontOfTarget: true } } } as any, host());

		expect(setReactionTime).not.toHaveBeenCalled();
		expect(setChannelClipDelay).not.toHaveBeenCalled();
	});

	it('applies the ones it does state', () => {
		applyBuild({ name: 'x', settings: { name: 'x', playerOptions: { reactionTimeMs: 250, channelClipDelayMs: 50 } } } as any, host());

		expect(setReactionTime).toHaveBeenCalledWith(250);
		expect(setChannelClipDelay).toHaveBeenCalledWith(50);
	});

	// A preset that means zero can say so; only `makePresetSettingsHelper` decides what "stated" is.
	it('applies a stated zero', () => {
		applyBuild({ name: 'x', settings: { name: 'x', playerOptions: { reactionTimeMs: 0 } } } as any, host());

		expect(setReactionTime).toHaveBeenCalledWith(0);
	});
});

describe('applyBuild reforge settings', () => {
	// Preset builds carry a partial; fromProto is the full-replace deserialization path.
	it('merges the preset instead of replacing every field', () => {
		const reforgeSettings = ReforgeSettingsProto.create({ useCustomEpValues: true });

		applyBuild({ name: 'P1', reforgeSettings }, host());

		expect(applyPresetSettings).toHaveBeenCalledWith(reforgeSettings);
		expect(fromProto).not.toHaveBeenCalled();
	});
});

describe('applyBuild item swap settings', () => {
	it("keeps a swap stone for the settings' swap weapon when the live swap set has none", async () => {
		const { player, host } = await makeIndividualSim();

		applyBuild(
			{
				name: 'Swap',
				settings: {
					name: 'Swap',
					consumables: ConsumesSpec.create({ swapMhImbueId: DENSE_SHARPENING_STONE_ID }),
					playerOptions: { enableItemSwap: true, itemSwap: swapSet({ mainHand: SWORD }) },
				},
			},
			host,
		);

		expect(player.itemSwapSettings.getItem(ItemSlot.ItemSlotMainHand)?.item.id).toBe(SWORD);
		expect(player.getConsumes().swapMhImbueId).toBe(DENSE_SHARPENING_STONE_ID);
	});
});
