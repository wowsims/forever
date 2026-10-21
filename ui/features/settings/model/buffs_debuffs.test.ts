import { Class, Stat } from '@generated/proto/common';
import { ActionId } from '@sim/proto/action_id';
import { describe, expect, it } from 'vitest';

import * as BuffDebuffInputs from './buffs_debuffs';
import { isMultiIconOption, type RenderableStatOptions } from './stat_options';

// The three registries interleave the generated rows with the hand-written ones. These sequences are
// the settings tab's layout, so a diff here is a UI change and wants a reason, not a re-snapshot.
describe('the buff registries', () => {
	// A group reads as its label followed by its members' labels.
	const labelsOf = (options: ReadonlyArray<RenderableStatOptions>) =>
		options.map(option =>
			isMultiIconOption(option) && option.members ? [option.config.label, option.members.map(member => member.config.label)] : option.config.label,
		);
	const groupWith = (options: ReadonlyArray<RenderableStatOptions>, config: unknown) =>
		options.find(option => isMultiIconOption(option) && option.members?.some(member => member.config === config));
	const memberOf = (options: ReadonlyArray<RenderableStatOptions>, config: unknown) => {
		const group = groupWith(options, config);
		return group && isMultiIconOption(group) ? group.members?.find(member => member.config === config) : undefined;
	};

	it('shows the party buffs in their settled order', () => {
		expect(labelsOf(BuffDebuffInputs.PARTY_BUFFS_CONFIG)).toEqual([
			'Blood Pact',
			'Devotion Aura',
			'Mana Spring Totem',
			'Mana Tide Totem',
			'Retribution Aura',
			'Concentration Aura',
			'Dark Inspiration',
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
			['settings_tab.raid_buffs.shadow_resistance', ['Prayer of Shadow Protection', 'Shadow Resistance Aura']],
			['settings_tab.raid_buffs.fire_resistance', ['Fire Resistance Aura', 'Fire Resistance Totem']],
			['settings_tab.raid_buffs.frost_resistance', ['Frost Resistance Aura', 'Frost Resistance Totem']],
			['settings_tab.raid_buffs.nature_resistance', ['Nature Resistance Totem', 'Aspect of the Wild']],
			'Strength of Earth Totem',
			'Grace of Air Totem',
			'Windfury Totem',
			'Flametongue Totem',
			['settings_tab.raid_buffs.attack_power', ['Battle Shout', 'Furious Howl']],
			'Trueshot Aura',
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
			'Curse of Weakness',
			['settings_tab.debuffs.major_armor_reduction', ['Sunder Armor', 'Expose Armor']],
			['settings_tab.debuffs.minor_armor_reduction', ['Faerie Fire', 'Curse of Recklessness', 'Dust Cloud']],
			'Armor Shatter',
			'Gift of Arthas',
			[
				'settings_tab.debuffs.attack_power_reduction',
				['Demoralizing Shout', 'Demoralizing Roar', 'Demoralizing Screech', 'Hex of Weakness', 'Vindication'],
			],
			['settings_tab.debuffs.attack_speed_reduction', ['Thunder Clap', 'Thunderfury']],
			['settings_tab.debuffs.hit_chance_reduction', ['Insect Swarm', 'Scorpid Sting']],
		]);
	});

	// The sim lets only the strongest of each set apply, so a set is one picker whose inputs are the generated ones.
	it('groups the buffs and debuffs that do not stack around the generated inputs themselves', () => {
		const attackPower = groupWith(BuffDebuffInputs.BUFFS_CONFIG, BuffDebuffInputs.BattleShout);
		expect(attackPower && isMultiIconOption(attackPower) ? attackPower.config.inputs : []).toEqual([
			BuffDebuffInputs.BattleShout,
			BuffDebuffInputs.FuriousHowl,
		]);
		expect(groupWith(BuffDebuffInputs.DEBUFFS_CONFIG, BuffDebuffInputs.SunderArmor)?.stats).toEqual([Stat.StatAttackPower]);
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
		expect(memberOf(BuffDebuffInputs.BUFFS_CONFIG, BuffDebuffInputs.BattleShout)?.ownerClass).toBe(Class.ClassWarrior);
		expect(BuffDebuffInputs.BUFFS_CONFIG[8].ownerClass).toBeUndefined();
	});
});
