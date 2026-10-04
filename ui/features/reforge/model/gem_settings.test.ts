import { ItemQuality, ItemSlot, Stat } from '@generated/proto/common';
import { CURRENT_PHASE, Phase } from '@sim/constants/other';
import { Stats } from '@sim/proto/stats';
import { ReforgeSettings } from '@sim/settings/reforge_settings';
import { createSimStore } from '@sim/state/sim_store';
import { describe, expect, it } from 'vitest';

const EP_STATS = [Stat.StatIntellect, Stat.StatSpellDamage, Stat.StatHitRating];

const makeSettings = (phase = CURRENT_PHASE) =>
	new ReforgeSettings({ sim: { store: createSimStore(), getPhase: () => phase }, storeKey: 1 } as any, {}, EP_STATS);

describe('gem optimizer settings', () => {
	it('defaults the three gem knobs the optimizer is driven by', () => {
		const settings = makeSettings();

		expect(settings.getMaxGemPhase()).toBe(CURRENT_PHASE);
		expect(settings.getMaxGemQuality()).toBe(ItemQuality.ItemQualityEpic);
		expect(settings.disableUniqueGems).toBe(false);
	});

	it('serializes the gem knobs and the spec’s gemmable stats onto ReforgeSettings', () => {
		const settings = makeSettings();
		settings.setMaxGemPhase(Phase.Phase2);
		settings.setMaxGemQuality(ItemQuality.ItemQualityRare);
		settings.setDisableUniqueGems(true);

		const proto = settings.toProto();

		expect(proto.maxGemPhase).toBe(Phase.Phase2);
		expect(proto.maxGemQuality).toBe(ItemQuality.ItemQualityRare);
		expect(proto.disableUniqueGems).toBe(true);
		expect(proto.epStats).toEqual(EP_STATS);
	});

	it('round-trips the gem knobs back out of a proto', () => {
		const source = makeSettings();
		source.setMaxGemPhase(Phase.Phase5);
		source.setMaxGemQuality(ItemQuality.ItemQualityUncommon);
		source.setDisableUniqueGems(true);

		const restored = makeSettings();
		restored.fromProto(source.toProto());

		expect(restored.getMaxGemPhase()).toBe(Phase.Phase5);
		expect(restored.getMaxGemQuality()).toBe(ItemQuality.ItemQualityUncommon);
		expect(restored.disableUniqueGems).toBe(true);
	});

	it('falls back to phase 1 and epic quality when a saved proto carries neither', () => {
		const settings = makeSettings();
		settings.fromProto({ ...settings.toProto(), maxGemPhase: 0, maxGemQuality: 0 });

		expect(settings.getMaxGemPhase()).toBe(Phase.Phase1);
		expect(settings.getMaxGemQuality()).toBe(ItemQuality.ItemQualityEpic);
	});

	it('applies the sim’s current phase, not the compiled-in one, when defaults are reapplied', () => {
		const settings = makeSettings(Phase.Phase4);
		settings.setMaxGemPhase(Phase.Phase1);
		settings.setDisableUniqueGems(true);
		settings.setFreezeItemSlots(true);
		settings.setFrozenItemSlots([ItemSlot.ItemSlotHead]);

		settings.applyDefaults();

		expect(settings.getMaxGemPhase()).toBe(Phase.Phase4);
		expect(settings.disableUniqueGems).toBe(false);
		expect(settings.freezeItemSlots).toBe(false);
		expect(settings.statCaps).toEqual(new Stats());
	});
});
