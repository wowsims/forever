import { Class, PseudoStat, Stat } from '@generated/proto/common';
import type { Player } from '@sim/player/player';
import { UnitStat } from '@sim/proto/stats';
import { fakeHost } from '@testing/sim/fake_host';
import { describe, expect, it } from 'vitest';

import { applyOwnerClassLabels, inDisplayOrder, type PickerStatOptions, relevantStatOptions, type RenderableStatOptions } from './stat_options';

const option = (stats: Array<Stat>) => ({ config: { label: stats.join('/') }, stats }) as unknown as PickerStatOptions;

const manaSpring = option([Stat.StatMP5]);
const manaTide = option([Stat.StatMP5]);
const untagged = option([]);
const strengthOfEarth = option([Stat.StatStrength]);
const fortitude = option([Stat.StatStamina]);
const options = [manaSpring, manaTide, untagged, strengthOfEarth, fortitude];

const host = (parts: { epStats?: Array<Stat>; displayStats?: Array<UnitStat>; include?: Array<unknown>; exclude?: Array<unknown> }) =>
	fakeHost({
		individualConfig: {
			epStats: parts.epStats ?? [],
			displayStats: parts.displayStats ?? [],
			includeBuffDebuffInputs: parts.include ?? [],
			excludeBuffDebuffInputs: parts.exclude ?? [],
		},
	});

describe('relevantStatOptions', () => {
	it('keeps an option tagged with an EP stat or a displayed stat, and every untagged option', () => {
		const shown = relevantStatOptions(options, host({ epStats: [Stat.StatMP5], displayStats: [UnitStat.fromStat(Stat.StatStamina)] }));
		expect(shown).toEqual([manaSpring, manaTide, untagged, fortitude]);
	});

	it('matches a displayed pseudo stat by its root rating, never by its own number', () => {
		const hitBuff = option([Stat.StatHitRating]);
		expect(PseudoStat.PseudoStatMeleeSpeedMultiplier).toBe(Stat.StatHitRating);
		const speed = host({ displayStats: [UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeSpeedMultiplier)] });
		expect(relevantStatOptions([hitBuff], speed)).toEqual([]);
		const hit = host({ displayStats: [UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeHitPercent)] });
		expect(relevantStatOptions([hitBuff], hit)).toEqual([hitBuff]);
	});

	it('includes by stat, the way every spec lists them', () => {
		expect(relevantStatOptions(options, host({ include: [Stat.StatStrength] }))).toEqual([untagged, strengthOfEarth]);
	});

	it('excludes by stat, the sentinel way the feral specs drop Windfury', () => {
		expect(relevantStatOptions(options, host({ epStats: [Stat.StatMP5], exclude: [Stat.StatMP5] }))).toEqual([untagged]);
	});

	it('includes and excludes a single input by its config, so one MP5 buff can go while the other stays', () => {
		const shown = relevantStatOptions(options, host({ epStats: [Stat.StatMP5], include: [fortitude.config], exclude: [manaTide.config, untagged.config] }));
		expect(shown).toEqual([manaSpring, fortitude]);
	});
});

const ownedOption = (label: string, ownerClass?: Class) => ({ config: { label }, stats: [], ownerClass }) as unknown as RenderableStatOptions;
const playerOf = (playerClass: Class) => ({ getClass: () => playerClass }) as unknown as Player<any>;

const battleShout = ownedOption('Battle Shout', Class.ClassWarrior);
const demoralizingShout = ownedOption('Demoralizing Shout', Class.ClassWarrior);
const arcaneBrilliance = ownedOption('Arcane Brilliance', Class.ClassMage);
const giftOfArthas = ownedOption('Gift of Arthas');
const unlabelledWarriorRow = { config: {}, stats: [], ownerClass: Class.ClassWarrior } as unknown as RenderableStatOptions;

describe('applyOwnerClassLabels', () => {
	it('marks the buffs the player casts itself as external, and returns every other row untouched', () => {
		const shown = applyOwnerClassLabels([battleShout, arcaneBrilliance, giftOfArthas], playerOf(Class.ClassWarrior));

		expect(shown.map(option => option.config.label)).toEqual(['Battle Shout (External)', 'Arcane Brilliance', 'Gift of Arthas']);
		expect(shown[1]).toBe(arcaneBrilliance);
		expect(shown[2]).toBe(giftOfArthas);
	});

	it('copies the relabelled row instead of renaming the shared config', () => {
		const shown = applyOwnerClassLabels([battleShout], playerOf(Class.ClassWarrior));

		expect(shown[0]).not.toBe(battleShout);
		expect(shown[0].config).not.toBe(battleShout.config);
		expect(battleShout.config.label).toBe('Battle Shout');
	});

	it('marks every row the class owns, not just the first', () => {
		const shown = applyOwnerClassLabels([battleShout, arcaneBrilliance, demoralizingShout], playerOf(Class.ClassWarrior));

		expect(shown.map(option => option.config.label)).toEqual(['Battle Shout (External)', 'Arcane Brilliance', 'Demoralizing Shout (External)']);
		expect(shown[1]).toBe(arcaneBrilliance);
	});

	it('leaves a row with no label alone, since there is nothing to mark', () => {
		const shown = applyOwnerClassLabels([unlabelledWarriorRow], playerOf(Class.ClassWarrior));

		expect(shown[0]).toBe(unlabelledWarriorRow);
	});

	it('runs after relevantStatOptions, so a spec excluding a row by its config still drops it', () => {
		const shown = applyOwnerClassLabels(
			relevantStatOptions([battleShout, arcaneBrilliance], host({ exclude: [battleShout.config] })),
			playerOf(Class.ClassWarrior),
		);

		expect(shown.map(option => option.config.label)).toEqual(['Arcane Brilliance']);
	});
});

describe('inDisplayOrder', () => {
	const prebuiltBattleShout = ownedOption('Battle Shout', Class.ClassWarrior);
	const prebuiltSunderArmor = ownedOption('Sunder Armor', Class.ClassWarrior);
	const handWritten = ownedOption('Blessing of Salvation');

	it('resolves a config to its prebuilt row and keeps the literal rows where they are written', () => {
		const composed = inDisplayOrder([prebuiltBattleShout, prebuiltSunderArmor], [prebuiltBattleShout.config, handWritten, prebuiltSunderArmor.config]);

		expect(composed).toEqual([prebuiltBattleShout, handWritten, prebuiltSunderArmor]);
		expect(composed[0]).toBe(prebuiltBattleShout);
	});

	it('reports a prebuilt row the display order never names, rather than dropping it off the tab', () => {
		expect(() => inDisplayOrder([prebuiltBattleShout, prebuiltSunderArmor], [prebuiltBattleShout.config])).toThrowError(
			'the display order leaves out Sunder Armor',
		);
	});

	it('reports a config with no prebuilt row', () => {
		expect(() => inDisplayOrder([prebuiltSunderArmor], [prebuiltBattleShout.config])).toThrowError(
			'the display order names Battle Shout, which no prebuilt row carries',
		);
	});
});
