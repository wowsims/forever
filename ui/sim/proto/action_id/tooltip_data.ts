import type { EquippedItem } from '../equipped_item';
import type { Gear } from '../gear';
import { hasItemTooltip, hasSpellTooltip, itemTooltipAnchor, spellTooltipAnchor } from '../tooltips';
import type { WowheadTooltipItemParams, WowheadTooltipSpellParams } from '../wowhead';
import { ActionId } from './index';

/** The tooltip props for an action's anchor: an item or spell with a tooltip of its own shows that, anything else wowhead's. */
export async function actionIdTooltipProps(actionId: ActionId, options: { useBuffAura?: boolean } = {}): Promise<Record<string, string>> {
	if (actionId.itemId) {
		if (await hasItemTooltip(actionId.itemId)) return itemTooltipAnchor(actionId.itemId, { randomSuffix: actionId.randomSuffixId });
	} else {
		const spellId = actionId.spellIdTooltipOverride || actionId.spellId;
		if (spellId && (await hasSpellTooltip(spellId))) return spellTooltipAnchor(spellId, { buff: options.useBuffAura });
	}
	return { 'data-wowhead': await actionIdWowheadTooltipData(actionId, options) };
}

/** The tooltip props for an equipped item's anchor, its own with the suffix and enchant it carries, or wowhead's. */
export async function equippedItemTooltipProps(gear: Gear, equippedItem: EquippedItem): Promise<Record<string, string>> {
	if (await hasItemTooltip(equippedItem.id)) {
		return itemTooltipAnchor(equippedItem.id, { randomSuffix: equippedItem.randomSuffix?.id, enchant: equippedItem.enchant?.effectId });
	}
	return { 'data-wowhead': await equippedItemWowheadTooltipData(gear, equippedItem) };
}

export function actionIdWowheadTooltipData(
	actionId: ActionId,
	params?: Omit<WowheadTooltipItemParams, 'itemId'> | Omit<WowheadTooltipSpellParams, 'spellId'>,
): Promise<string> {
	return actionId.itemId
		? ActionId.makeItemTooltipData(actionId.itemId, params)
		: ActionId.makeSpellTooltipData(actionId.spellIdTooltipOverride || actionId.spellId, params);
}

export function equippedItemWowheadTooltipData(gear: Gear, equippedItem: EquippedItem): Promise<string> {
	const enchantIds = [equippedItem.enchant?.effectId].filter((id): id is number => id !== undefined);
	return actionIdWowheadTooltipData(equippedItem.asActionId(), {
		itemLevel: Number(equippedItem.ilvl),
		enchantIds: enchantIds,
		randomEnchantmentId: equippedItem.randomSuffix?.id,
		setPieceIds: gear
			.asArray()
			.filter(ei => ei != null)
			.map(ei => ei!.item.id),
	});
}
