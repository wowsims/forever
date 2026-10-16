import { Class } from '@generated/proto/common';
import { ActionId } from '@sim/proto/action_id';
import { describe, expect, it } from 'vitest';

import * as BuffDebuffInputs from './buffs_debuffs';
import type { RenderableStatOptions } from './stat_options';

// The three registries interleave the generated rows with the hand-written ones. These sequences are
// the settings tab's layout, so a diff here is a UI change and wants a reason, not a re-snapshot.
describe('the buff registries', () => {
	const labelsOf = (options: ReadonlyArray<RenderableStatOptions>) => options.map(option => option.config.label);

	it('shows the party buffs in their settled order', () => {
		expect(labelsOf(BuffDebuffInputs.PARTY_BUFFS_CONFIG)).toEqual([
			'Blood Pact',
			'Devotion Aura',
			'Mana Spring Totem',
			'Mana Tide Totem',
			'Retribution Aura',
			'Concentration Aura',
			'Atiesh - Mage',
			'Atiesh - Warlock',
			'Atiesh - Druid',
			'Atiesh - Priest',
			'Darkspear',
			"Admiral's Hat",
			'Stoneskin Totem',
		]);
	});

	it('shows the raid and individual buffs interleaved in their settled order', () => {
		expect(labelsOf(BuffDebuffInputs.BUFFS_CONFIG)).toEqual([
			'Arcane Brilliance',
			'Greater Blessing of Kings',
			'Prayer of Spirit',
			'Gift of the Wild',
			'Thorns',
			'Prayer of Fortitude',
			'Greater Blessing of Might',
			'Greater Blessing of Wisdom',
			'Mass Salvation',
			'Greater Blessing of Light',
			'Prayer of Shadow Protection',
			'Fire Resistance Aura',
			'Frost Resistance Aura',
			'Shadow Resistance Aura',
			'Fire Resistance Totem',
			'Frost Resistance Totem',
			'Nature Resistance Totem',
			'Strength of Earth Totem',
			'Grace of Air Totem',
			'Windfury Totem',
			'Flametongue Totem',
			'Battle Shout',
			'Trueshot Aura',
			'Furious Howl',
			'Aspect of the Wild',
			'Leader of the Pack',
			'Moonkin Aura',
			'Innervates',
			'Power Infusions',
		]);
	});

	it('shows the debuffs in their settled order', () => {
		expect(labelsOf(BuffDebuffInputs.DEBUFFS_CONFIG)).toEqual([
			"Hunter's Mark",
			'Judgement of the Crusader',
			'Judgement of Light',
			'Judgement of Wisdom',
			'Curse of the Elements',
			'Curse of Recklessness',
			'Curse of Weakness',
			'Faerie Fire',
			'Dust Cloud',
			'Expose Armor',
			'Sunder Armor',
			'Armor Shatter',
			'Gift of Arthas',
			'Demoralizing Roar',
			'Demoralizing Shout',
			'Hex of Weakness',
			'Demoralizing Screech',
			'Vindication',
			'Thunder Clap',
			'Thunderfury',
			'Insect Swarm',
			'Scorpid Sting',
		]);
	});

	it('puts the hand-written inputs themselves at those positions, not lookalikes', () => {
		expect(BuffDebuffInputs.BUFFS_CONFIG[8].config).toBe(BuffDebuffInputs.GreaterBlessingOfSalvation);
	});

	// Off, the shout, and the shout from a warrior wearing three pieces of Battlegear of Wrath, whose
	// set spell 23563 is the improved corner's icon: one picker rather than a second boolean next to it.
	it('offers Battle Shout as a single three-state icon', () => {
		expect(BuffDebuffInputs.BattleShout.states).toBe(3);
		expect(BuffDebuffInputs.BattleShout.improvedId?.equals(ActionId.fromSpellId(23563))).toBe(true);
	});

	it('gives the generated rows their owner class, so the settings tab can mark them external', () => {
		const battleShout = BuffDebuffInputs.BUFFS_CONFIG.find(option => option.config === BuffDebuffInputs.BattleShout);
		expect(battleShout?.ownerClass).toBe(Class.ClassWarrior);
		expect(BuffDebuffInputs.BUFFS_CONFIG[8].ownerClass).toBeUndefined();
	});
});
