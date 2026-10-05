import { APLRotation, APLRotation_Type } from '@generated/proto/apl';
import { ActionID, EquipmentSpec, HandType, ItemQuality, ItemSlot, ItemSpec, ItemType, WeaponType } from '@generated/proto/common';
import { UIEnchant, UIItem } from '@generated/proto/ui';

// Ids and icons are real rows of the Forever item database, so a story that loads tooltips shows the
// right one.

/** A head (Lionheart Helm) and a main hand (Thunderfury with Crusader); every other slot is empty. */
export const makeEquipmentSpec = (
	items: Partial<Record<ItemSlot, Partial<ItemSpec>>> = {
		[ItemSlot.ItemSlotHead]: { id: 12640 },
		[ItemSlot.ItemSlotMainHand]: { id: 19019, enchant: 1900 },
	},
): EquipmentSpec => {
	const slots = Object.keys(items).map(Number);
	const length = slots.length ? Math.max(...slots) + 1 : 0;
	return EquipmentSpec.create({ items: Array.from({ length }, (_, slot) => ItemSpec.create(items[slot as ItemSlot])) });
};

export const makeUIItem = (overrides: Partial<UIItem> = {}): UIItem =>
	UIItem.create({
		id: 19019,
		name: 'Thunderfury, Blessed Blade of the Windseeker',
		icon: 'inv_sword_39',
		type: ItemType.ItemTypeWeapon,
		weaponType: WeaponType.WeaponTypeSword,
		handType: HandType.HandTypeOneHand,
		weaponDamageMin: 44,
		weaponDamageMax: 115,
		weaponSpeed: 1.9,
		ilvl: 80,
		phase: 1,
		quality: ItemQuality.ItemQualityLegendary,
		unique: true,
		...overrides,
	});

export const makeUIEnchant = (overrides: Partial<UIEnchant> = {}): UIEnchant =>
	UIEnchant.create({
		effectId: 1900,
		itemId: 16252,
		spellId: 20007,
		name: 'Enchant Weapon - Crusader',
		icon: 'inv_misc_note_01',
		type: ItemType.ItemTypeWeapon,
		quality: ItemQuality.ItemQualityUncommon,
		...overrides,
	});

const castSpell = (spellId: number) => ({
	hide: false,
	notes: '',
	action: { action: { oneofKind: 'castSpell' as const, castSpell: { spellId: ActionID.create({ rawId: { oneofKind: 'spellId', spellId } }) } } },
});

/** An APL rotation casting Bloodthirst, then Whirlwind, then Heroic Strike. */
export const makeAplRotation = (overrides: Partial<APLRotation> = {}): APLRotation =>
	APLRotation.create({
		type: APLRotation_Type.TypeAPL,
		priorityList: [castSpell(23894), castSpell(1680), castSpell(25286)],
		...overrides,
	});
