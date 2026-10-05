import { ReforgeSettings as ReforgeSettingsProto, StatCapType } from '@generated/proto/api';
import { ItemSlot, Stat } from '@generated/proto/common';
import { StatCap, Stats } from '@sim/proto/stats';
import { ReforgeSettings } from '@sim/settings/reforge_settings';
import { createSimStore } from '@sim/state/sim_store';
import { describe, expect, it } from 'vitest';

// A spec with soft caps, the way warrior/dps has them: applyDefaults() turns
// useSoftCapBreakpoints on because the spec ships breakpoints.
const withSoftCaps = () => {
	const defaults = {
		statCaps: new Stats().withStat(Stat.StatArmorPenetration, 1400),
		softCapBreakpoints: [StatCap.fromStat(Stat.StatArmorPenetration, { breakpoints: [1400], capType: StatCapType.TypeSoftCap, postCapEPs: [0] })],
	};
	const settings = new ReforgeSettings({ sim: { store: createSimStore() }, storeKey: 1 } as any, defaults as any);
	settings.applyDefaults();
	return settings;
};

describe('ReforgeSettings.applyPreset', () => {
	// A partial preset: one field set, everything else left at its proto zero value.
	const customEpOnly = ReforgeSettingsProto.create({ useCustomEpValues: true });

	it('applies the fields the preset sets', () => {
		const settings = withSoftCaps();
		settings.setUseCustomEPValues(false);

		settings.applyPreset(customEpOnly);

		expect(settings.useCustomEPValues).toBe(true);
	});

	// `fromProto` is full-replace, so a preset that only names one field must not switch the
	// spec's soft-cap breakpoints off.
	it('leaves a field the preset omits at its spec default', () => {
		const settings = withSoftCaps();
		expect(settings.useSoftCapBreakpoints).toBe(true);

		settings.applyPreset(customEpOnly);

		expect(settings.useSoftCapBreakpoints).toBe(true);
	});

	// `create()` seeds repeated fields to `[]`, so a truthiness guard here fires on every preset.
	it('leaves the user frozen slots a preset does not mention', () => {
		const settings = withSoftCaps();
		settings.setFreezeItemSlots(true);
		settings.setFrozenItemSlots([ItemSlot.ItemSlotHead, ItemSlot.ItemSlotChest]);

		settings.applyPreset(customEpOnly);

		expect([...settings.frozenItemSlots]).toEqual([ItemSlot.ItemSlotHead, ItemSlot.ItemSlotChest]);
	});

	it('still lets a preset set frozen slots', () => {
		const settings = withSoftCaps();

		settings.applyPreset(ReforgeSettingsProto.create({ frozenItemSlots: [ItemSlot.ItemSlotHands] }));

		expect([...settings.frozenItemSlots]).toEqual([ItemSlot.ItemSlotHands]);
	});

	it('is not fromProto: fromProto does reset the omitted field', () => {
		const settings = withSoftCaps();

		settings.fromProto(customEpOnly);

		expect(settings.useSoftCapBreakpoints).toBe(false);
	});
});
