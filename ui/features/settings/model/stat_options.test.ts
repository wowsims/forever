import { inputPicked } from '@features/inputs/components/ConnectedIconEnumPicker/utils';
import { Class, PseudoStat, Stat } from '@generated/proto/common';
import type { Player } from '@sim/player/player';
import { UnitStat } from '@sim/proto/stats';
import { describe, expect, it } from 'vitest';

import {
	applyOwnerClassLabels,
	type IconPickerStatOption,
	inDisplayOrder,
	isMultiIconOption,
	type PickerStatOptions,
	relevantStatOptions,
	type RenderableStatOptions,
	statOptionGroup,
	statOptionScope,
} from './stat_options';

/** Stands in for the player: one value per option label. */
type Picks = Record<string, number | boolean>;
const playerWith = (picks: Picks) => ({ picks }) as unknown as Player<any>;
const pickOf = (player: Player<any>, label: string) => (player as unknown as { picks: Picks }).picks[label] ?? 0;

const option = (stats: Array<Stat>, label = stats.join('/') || 'untagged') =>
	({ config: { label, getValue: (player: Player<any>) => pickOf(player, label) }, stats }) as unknown as PickerStatOptions;

const manaSpring = option([Stat.StatMP5], 'Mana Spring');
const manaTide = option([Stat.StatMP5], 'Mana Tide');
const untagged = option([]);
const strengthOfEarth = option([Stat.StatStrength], 'Strength of Earth');
const fortitude = option([Stat.StatStamina], 'Fortitude');
const options = [manaSpring, manaTide, untagged, strengthOfEarth, fortitude];

// The icon config a fixture row carries, as a group names its members.
const iconConfig = (row: PickerStatOptions | RenderableStatOptions) => row.config as IconPickerStatOption['config'];

const lists = (parts: { include?: Array<unknown>; exclude?: Array<unknown> } = {}) => ({
	includeBuffDebuffInputs: parts.include ?? [],
	excludeBuffDebuffInputs: parts.exclude ?? [],
});

const specStats = (parts: { epStats?: Array<Stat>; displayStats?: Array<UnitStat>; include?: Array<unknown>; exclude?: Array<unknown> } = {}) => ({
	epStats: parts.epStats ?? [],
	displayStats: parts.displayStats ?? [],
	...lists(parts),
});

const offered = (shown: ReadonlyArray<PickerStatOptions>) => shown.filter(row => !row.config.keptBy);

// The kept mark an adapter shows, for the value it reads: the input's own, or for a multi icon whether any input is picked.
const marked = (config: PickerStatOptions['config'], player: Player<any>) =>
	!!config.keptBy?.(player, ('inputs' in config ? inputPicked(config, player) : config.getValue(player)) as never);

describe('relevantStatOptions', () => {
	it('offers an option tagged with a shown stat and every untagged option, as the same objects', () => {
		const shown = relevantStatOptions(options, lists(), new Set([Stat.StatMP5, Stat.StatStamina]));

		expect(offered(shown)).toEqual([manaSpring, manaTide, untagged, fortitude]);
		offered(shown).forEach((row, index) => expect(row).toBe([manaSpring, manaTide, untagged, fortitude][index]));
	});

	it('returns an option whose stats are hidden in its place, shown and marked only while it is picked', () => {
		const shown = relevantStatOptions(options, lists(), new Set([Stat.StatMP5]));
		const kept = shown[3];

		expect(kept.stats).toEqual([Stat.StatStrength]);
		expect(kept.config).not.toBe(strengthOfEarth.config);
		expect(kept.config.showWhen!(playerWith({}))).toBe(false);
		expect(marked(kept.config, playerWith({}))).toBe(false);
		expect(kept.config.showWhen!(playerWith({ 'Strength of Earth': true }))).toBe(true);
		expect(marked(kept.config, playerWith({ 'Strength of Earth': true }))).toBe(true);
		expect(strengthOfEarth.config.keptBy).toBeUndefined();
		expect(strengthOfEarth.config.showWhen).toBeUndefined();
	});

	it("keeps the option's own showWhen, so a picked option the spec hides stays hidden", () => {
		const tankOnly = { ...strengthOfEarth, config: { ...strengthOfEarth.config, showWhen: () => false } } as PickerStatOptions;
		const [kept] = relevantStatOptions([tankOnly], lists(), new Set());

		expect(kept.config.showWhen!(playerWith({ 'Strength of Earth': true }))).toBe(false);
	});

	it('reads an icon enum as picked off its zero value, and a multi icon as picked when any input is', () => {
		const enumOption = {
			stats: [Stat.StatStrength],
			config: {
				values: [],
				zeroValue: 'none',
				equals: (a: string, b: string) => a === b,
				getValue: (player: Player<any>) => pickOf(player, 'enum') || 'none',
			},
		} as unknown as PickerStatOptions;
		const multiOption = { stats: [Stat.StatStrength], config: { inputs: [manaSpring.config, manaTide.config] } } as unknown as PickerStatOptions;
		const [keptEnum, keptMulti] = relevantStatOptions([enumOption, multiOption], lists(), new Set());

		expect(keptEnum.config.showWhen!(playerWith({}))).toBe(false);
		expect(marked(keptEnum.config, playerWith({}))).toBe(false);
		expect(keptEnum.config.showWhen!(playerWith({ enum: 'greater' as unknown as number }))).toBe(true);
		expect(marked(keptEnum.config, playerWith({ enum: 'greater' as unknown as number }))).toBe(true);
		expect(keptMulti.config.showWhen!(playerWith({}))).toBe(false);
		expect(marked(keptMulti.config, playerWith({}))).toBe(false);
		expect(keptMulti.config.showWhen!(playerWith({ 'Mana Tide': 2 }))).toBe(true);
		expect(marked(keptMulti.config, playerWith({ 'Mana Tide': 2 }))).toBe(true);
	});

	it('offers only the picked inputs of a kept multi icon, each marked kept, and leaves the shared inputs as they were', () => {
		const multiOption = { stats: [Stat.StatMP5], config: { inputs: [manaSpring.config, manaTide.config] } } as unknown as PickerStatOptions;
		const [kept] = relevantStatOptions([multiOption], lists(), new Set());
		const [spring, tide] = (kept.config as unknown as { inputs: Array<PickerStatOptions['config']> }).inputs;
		const player = playerWith({ 'Mana Tide': 1 });

		expect(spring.showWhen!(player)).toBe(false);
		expect(marked(spring, player)).toBe(false);
		expect(tide.showWhen!(player)).toBe(true);
		expect(marked(tide, player)).toBe(true);
		expect(manaSpring.config.showWhen).toBeUndefined();
		expect(manaSpring.config.keptBy).toBeUndefined();
	});

	it('leaves a multi icon with a shown stat, and its inputs, untouched', () => {
		const multiOption = { stats: [Stat.StatMP5], config: { inputs: [manaSpring.config, manaTide.config] } } as unknown as PickerStatOptions;

		expect(relevantStatOptions([multiOption], lists(), new Set([Stat.StatMP5]))[0]).toBe(multiOption);
	});

	it('returns the same rows in the same order at every filter setting, so the pickers keyed by position keep their option', () => {
		const settings = [new Set<Stat>(), new Set([Stat.StatMP5]), new Set([Stat.StatStrength, Stat.StatStamina])];

		for (const shown of settings) {
			expect(relevantStatOptions(options, lists(), shown).map(row => row.config.label)).toEqual(options.map(row => row.config.label));
		}
	});

	it('offers an input its config includes, whatever the filter shows', () => {
		const shown = relevantStatOptions(options, lists({ include: [fortitude.config] }), new Set());

		expect(shown[4]).toBe(fortitude);
	});

	it('drops an input its config excludes outright, instead of keeping it while picked', () => {
		const shown = relevantStatOptions(options, lists({ exclude: [manaTide.config, untagged.config] }), new Set([Stat.StatMP5]));

		expect(shown.map(row => row.config.label)).toEqual(['Mana Spring', 'Strength of Earth', 'Fortitude']);
	});

	it('drops every option tagged with an excluded stat at every filter setting, and leaves that stat out of the universe', () => {
		const strengthAgility = option([Stat.StatStrength, Stat.StatAgility], 'Strength and Agility');
		const agilityOnly = option([Stat.StatAgility], 'Agility');
		const tagged = [strengthAgility, agilityOnly, strengthOfEarth];
		const settings = [new Set<Stat>(), new Set([Stat.StatStrength]), new Set([Stat.StatAgility]), new Set([Stat.StatStrength, Stat.StatAgility])];

		for (const shown of settings) {
			expect(relevantStatOptions(tagged, lists({ exclude: [Stat.StatAgility] }), shown).map(row => row.config.label)).toEqual(['Strength of Earth']);
		}
		expect(statOptionScope(tagged, specStats({ exclude: [Stat.StatAgility] })).universe).toEqual([UnitStat.fromStat(Stat.StatStrength)]);
	});
});

describe('relevantStatOptions on a group', () => {
	const flametongue = option([Stat.StatAttackPower], 'Flametongue');
	const battleShout = option([Stat.StatAttackPower], 'Battle Shout');
	const furiousHowl = option([Stat.StatAttackPower], 'Furious Howl');
	const [attackPower] = inDisplayOrder([battleShout, furiousHowl], [statOptionGroup('Attack Power', [iconConfig(battleShout), iconConfig(furiousHowl)])]);
	const inputLabels = (row: PickerStatOptions | undefined) => (row && isMultiIconOption(row) ? row.config.inputs.map(input => input.label) : []);

	it('drops the members the spec excludes and keeps the rest under the group label', () => {
		const [shown] = relevantStatOptions([attackPower], lists({ exclude: [furiousHowl.config] }), new Set([Stat.StatAttackPower]));

		expect(shown.config.label).toBe('Attack Power');
		expect(inputLabels(shown)).toEqual(['Battle Shout']);
	});

	it('drops the group with its last member', () => {
		expect(
			relevantStatOptions([attackPower, flametongue], lists({ exclude: [battleShout.config, furiousHowl.config] }), new Set([Stat.StatAttackPower])),
		).toEqual([flametongue]);
	});

	it('offers the group when the spec includes one of its members', () => {
		const [shown] = relevantStatOptions([attackPower], lists({ include: [furiousHowl.config] }), new Set());

		expect(shown).toBe(attackPower);
	});
});

describe('statOptionScope', () => {
	const defaultStats = (parts: Parameters<typeof specStats>[0]) => statOptionScope(options, specStats(parts)).defaults.map(stat => stat.getStat());

	it("has every stat tagged on the section's options once, in stat order, minus the excluded stats, as its universe", () => {
		const { universe } = statOptionScope(options, specStats({ exclude: [Stat.StatMP5, manaSpring.config] }));

		expect(universe).toEqual([UnitStat.fromStat(Stat.StatStrength), UnitStat.fromStat(Stat.StatStamina)]);
	});

	it('defaults to the EP stats, the displayed stats and the stats the spec includes, and never an included config', () => {
		const defaults = defaultStats({
			epStats: [Stat.StatMP5],
			displayStats: [UnitStat.fromStat(Stat.StatStamina)],
			include: [Stat.StatStrength, fortitude.config],
		});

		expect(new Set(defaults)).toEqual(new Set([Stat.StatStamina, Stat.StatMP5, Stat.StatStrength]));
	});

	it('takes a displayed pseudo stat into the defaults by its root rating, never by its own number', () => {
		expect(PseudoStat.PseudoStatMeleeSpeedMultiplier).toBe(Stat.StatHitRating);
		expect(defaultStats({ displayStats: [UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeSpeedMultiplier)] })).toEqual([]);
		expect(defaultStats({ displayStats: [UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeHitPercent)] })).toEqual([Stat.StatHitRating]);
	});
});

const ownedOption = (label: string, ownerClass?: Class) => ({ config: { label }, stats: [], ownerClass }) as unknown as RenderableStatOptions;
const playerOf = (playerClass: Class) => ({ getClass: () => playerClass }) as unknown as Player<any>;

const battleShout = ownedOption('Battle Shout', Class.ClassWarrior);
const demoralizingShout = ownedOption('Demoralizing Shout', Class.ClassWarrior);
const arcaneIntellect = ownedOption('Arcane Intellect', Class.ClassMage);
const giftOfArthas = ownedOption('Gift of Arthas');
const unlabelledWarriorRow = { config: {}, stats: [], ownerClass: Class.ClassWarrior } as unknown as RenderableStatOptions;

describe('applyOwnerClassLabels', () => {
	it('marks the buffs the player casts itself as external, and returns every other row untouched', () => {
		const shown = applyOwnerClassLabels([battleShout, arcaneIntellect, giftOfArthas], playerOf(Class.ClassWarrior));

		expect(shown.map(option => option.config.label)).toEqual(['settings_tab.external_label', 'Arcane Intellect', 'Gift of Arthas']);
		expect(shown[1]).toBe(arcaneIntellect);
		expect(shown[2]).toBe(giftOfArthas);
	});

	it('copies the relabelled row instead of renaming the shared config', () => {
		const shown = applyOwnerClassLabels([battleShout], playerOf(Class.ClassWarrior));

		expect(shown[0]).not.toBe(battleShout);
		expect(shown[0].config).not.toBe(battleShout.config);
		expect(battleShout.config.label).toBe('Battle Shout');
	});

	it('marks every row the class owns, not just the first', () => {
		const shown = applyOwnerClassLabels([battleShout, arcaneIntellect, demoralizingShout], playerOf(Class.ClassWarrior));

		expect(shown.map(option => option.config.label)).toEqual(['settings_tab.external_label', 'Arcane Intellect', 'settings_tab.external_label']);
		expect(shown[1]).toBe(arcaneIntellect);
	});

	it('leaves a row with no label alone, since there is nothing to mark', () => {
		const shown = applyOwnerClassLabels([unlabelledWarriorRow], playerOf(Class.ClassWarrior));

		expect(shown[0]).toBe(unlabelledWarriorRow);
	});

	it("marks a group's owned members, not the group, and returns a group with none untouched", () => {
		const [attackPower] = inDisplayOrder(
			[battleShout, arcaneIntellect],
			[statOptionGroup('Group', [iconConfig(battleShout), iconConfig(arcaneIntellect)])],
		);
		const [shown] = applyOwnerClassLabels([attackPower], playerOf(Class.ClassWarrior));

		expect(shown.config.label).toBe('Group');
		expect(isMultiIconOption(shown) && shown.config.inputs.map(input => input.label)).toEqual(['settings_tab.external_label', 'Arcane Intellect']);
		expect(applyOwnerClassLabels([attackPower], playerOf(Class.ClassPriest))[0]).toBe(attackPower);
	});

	it('runs after relevantStatOptions, so a spec excluding a row by its config still drops it', () => {
		const shown = applyOwnerClassLabels(
			relevantStatOptions([battleShout, arcaneIntellect], lists({ exclude: [battleShout.config] }), new Set()),
			playerOf(Class.ClassWarrior),
		);

		expect(shown.map(option => option.config.label)).toEqual(['Arcane Intellect']);
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

	it("builds a group from the prebuilt rows it names, with every member's stats, and counts its members as placed", () => {
		const shout = { ...prebuiltBattleShout, stats: [Stat.StatAttackPower] } as RenderableStatOptions;
		const sunder = { ...prebuiltSunderArmor, stats: [Stat.StatAttackPower, Stat.StatArmor] } as RenderableStatOptions;
		const [group] = inDisplayOrder([shout, sunder], [statOptionGroup('Warrior', [iconConfig(shout), iconConfig(sunder)])]);

		expect(group.config.label).toBe('Warrior');
		expect(group.stats).toEqual([Stat.StatAttackPower, Stat.StatArmor]);
		expect(isMultiIconOption(group) && group.members).toEqual([shout, sunder]);
	});

	it('reports a group member with no prebuilt row', () => {
		expect(() =>
			inDisplayOrder([prebuiltSunderArmor], [statOptionGroup('Warrior', [iconConfig(prebuiltSunderArmor), iconConfig(prebuiltBattleShout)])]),
		).toThrowError('the display order names Battle Shout, which no prebuilt row carries');
	});

	it('reports a config with no prebuilt row', () => {
		expect(() => inDisplayOrder([prebuiltSunderArmor], [prebuiltBattleShout.config])).toThrowError(
			'the display order names Battle Shout, which no prebuilt row carries',
		);
	});
});
