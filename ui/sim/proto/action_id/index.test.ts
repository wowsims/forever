import { ItemSlot, OtherAction } from '@generated/proto/common';
import i18n from '@i18n/config';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { Database } from '../database';
import { ActionId } from '.';

afterEach(() => vi.restoreAllMocks());

const filledName = async (actionId: ActionId, name: string) => {
	vi.spyOn(ActionId, 'getTooltipData').mockImplementation(async () => ({ name, icon: '' }) as never);
	return (await actionId.fill()).name;
};

describe('ActionId.fill', () => {
	const mainHand = i18n.t('common.action_tags.main_hand');
	const offHand = i18n.t('common.action_tags.off_hand');

	it("names a weapon hand's copy of a proc by the item slot it is tagged with", async () => {
		expect(await filledName(ActionId.fromSpellId(20007, 0, ItemSlot.ItemSlotMainHand), 'Holy Strength')).toBe(`Holy Strength (${mainHand})`);
		expect(await filledName(ActionId.fromSpellId(28093, 0, ItemSlot.ItemSlotOffHand), 'Lightning Speed')).toBe(`Lightning Speed (${offHand})`);
		expect(await filledName(ActionId.fromItemId(20749, ItemSlot.ItemSlotOffHand), 'Brilliant Wizard Oil')).toBe(`Brilliant Wizard Oil (${offHand})`);
	});

	it("names the auto attack's swings by their item slot and an extra attack by the spell that granted it", async () => {
		const spellNames = vi.spyOn(Database, 'getSpellIconData').mockImplementation(async id => ({ name: id === 3391 ? 'Thrash' : '' }) as never);
		const attack = (tag: number) => ActionId.fromOtherId(OtherAction.OtherActionAttack, tag).fill();
		const name = i18n.t('common.action_names.attack');
		expect((await attack(ItemSlot.ItemSlotMainHand)).name).toBe(`${name} (${mainHand})`);
		expect((await attack(ItemSlot.ItemSlotOffHand)).name).toBe(`${name} (${offHand})`);
		expect(spellNames).not.toHaveBeenCalled();
		expect((await attack(3391)).name).toBe(`${name} (Thrash)`);
	});

	it("names a strike's hand beside its untagged cast", async () => {
		expect(await filledName(ActionId.fromSpellId(1241584, 0, 0), 'Mutilate')).toBe(`Mutilate (${i18n.t('common.action_tags.cast')})`);
		expect(await filledName(ActionId.fromSpellId(1241584, 0, ItemSlot.ItemSlotMainHand), 'Mutilate')).toBe(`Mutilate (${mainHand})`);
		expect(await filledName(ActionId.fromSpellId(1241584, 0, ItemSlot.ItemSlotOffHand), 'Mutilate')).toBe(`Mutilate (${offHand})`);
		expect(await filledName(ActionId.fromSpellId(1680, 0, ItemSlot.ItemSlotOffHand), 'Whirlwind')).toBe(`Whirlwind (${offHand})`);
	});

	it('names what the sim means by a tag that is not a hand', async () => {
		const weapons = (...types: string[]) => types.map(type => i18n.t(`common.weapon_types.${type}` as never)).join('/');
		const tag = (key: string) => i18n.t(`common.action_tags.${key}` as never);
		const cases: Array<[ActionId, string, string]> = [
			[ActionId.fromSpellId(1290261, 0, 1), 'Weaponmaster', `Weaponmaster (${weapons('axe', 'polearm')})`],
			[ActionId.fromSpellId(1290261, 0, 2), 'Weaponmaster', `Weaponmaster (${weapons('mace', 'staff')})`],
			[ActionId.fromSpellId(1290261, 0, 3), 'Weaponmaster', `Weaponmaster (${weapons('sword')})`],
			[ActionId.fromSpellId(20423, 0, 1), 'Seal of Fury', `Seal of Fury (${tag('shield')})`],
			[ActionId.fromSpellId(10901, 0, 1), 'Power Word: Shield', `Power Word: Shield (${tag('shield')})`],
			[ActionId.fromSpellId(440667, 0, 1), 'Vindication', `Vindication (${tag('self')})`],
			[ActionId.fromSpellId(440667, 0, -1), 'Vindication', `Vindication (${tag('external')})`],
			[ActionId.fromSpellId(11708, 0, 1), 'Curse of Weakness', `Curse of Weakness (${tag('amplified')})`],
			[ActionId.fromSpellId(1225228, 0, 1), 'Bane of Havoc', `Bane of Havoc (${tag('damage')})`],
			[ActionId.fromSpellId(16979, 0, 1), 'Feral Charge (Bear)', `Feral Charge (Bear) (${i18n.t('common.action_names.moving')})`],
			[ActionId.fromSpellId(23266, 0, 2), 'Fiery Aura', `Fiery Aura (${tag('hit')})`],
			[ActionId.fromItemId(10646, 1), 'Goblin Sapper Charge', `Goblin Sapper Charge (${tag('self')})`],
			[ActionId.fromItemId(12662, 1), 'Demonic Rune', `Demonic Rune (${tag('self')})`],
			[ActionId.fromItemId(250951, 2), 'Caustic Smog Potion', `Caustic Smog Potion (${tag('dot')})`],
		];
		for (const [actionId, baseName, name] of cases) {
			expect(await filledName(actionId, baseName)).toBe(name);
		}
	});

	it('names the hand of an aura read from the combat log', async () => {
		const actionId = ActionId.fromLogString(`Aura gained: {SpellID: 20007, Tag: ${ItemSlot.ItemSlotMainHand}}`);
		expect(await filledName(actionId, 'Holy Strength')).toBe(`Holy Strength (${mainHand})`);
	});

	it('names no hand for any other tag', async () => {
		expect(await filledName(ActionId.fromSpellId(1, 0, 1), 'Unnamed Spell')).toBe('Unnamed Spell (??)');
		expect(await filledName(ActionId.fromSpellId(1, 0, 2), 'Unnamed Spell')).toBe('Unnamed Spell (??)');
		expect(await filledName(ActionId.fromSpellId(1, 0, -1), 'Unnamed Spell')).toBe(`Unnamed Spell (${i18n.t('common.action_tags.external')})`);
	});
});
