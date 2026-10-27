import { HandType, ItemSlot, ItemSwap, ItemType, WeaponType } from '@generated/proto/common';
import { UIDatabase, UIItem } from '@generated/proto/ui';
import type { Player } from '@sim/player/player';
import { makeEquipmentSpec } from '@testing/storybook/fixtures/protos';

// Two one-hand weapons, one sharp and one blunt, for the stone imbues to follow.
export const SWORD = 1;
export const MACE = 2;

const oneHand = (id: number, weaponType: WeaponType) =>
	UIItem.create({ id, type: ItemType.ItemTypeWeapon, handType: HandType.HandTypeOneHand, weaponType, scalingOptions: { 0: { ilvl: 60 } } });

export const weaponsDatabase = () => UIDatabase.create({ items: [oneHand(SWORD, WeaponType.WeaponTypeSword), oneHand(MACE, WeaponType.WeaponTypeMace)] });

// An item swap set indexed by slot, as the saved settings write it.
export const swapSet = (weapons: { mainHand?: number; offHand?: number }) =>
	ItemSwap.create({
		items: makeEquipmentSpec({
			[ItemSlot.ItemSlotMainHand]: { id: weapons.mainHand },
			[ItemSlot.ItemSlotOffHand]: { id: weapons.offHand },
		}).items,
	});

export const swapWeaponIds = (player: Player<any>) =>
	[ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand].map(slot => player.itemSwapSettings.getItem(slot)?.item.id);
