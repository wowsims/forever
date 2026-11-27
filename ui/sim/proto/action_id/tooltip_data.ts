import type { EquippedItem } from '../equipped_item';
import { hasItemTooltip, hasSpellTooltip, itemTooltipAnchor, spellTooltipAnchor } from '../tooltips';
import type { ActionId } from './index';

/** The tooltip props for an action's anchor: the item's or spell's own tooltip, or none for one the tooltips do not hold. */
export async function actionIdTooltipProps(actionId: ActionId, options: { useBuffAura?: boolean } = {}): Promise<Record<string, string>> {
	if (actionId.itemId) return (await hasItemTooltip(actionId.itemId)) ? itemTooltipAnchor(actionId.itemId, { randomSuffix: actionId.randomSuffixId }) : {};
	const spellId = actionId.spellIdTooltipOverride || actionId.spellId;
	return spellId && (await hasSpellTooltip(spellId)) ? spellTooltipAnchor(spellId, { buff: options.useBuffAura }) : {};
}

/** The tooltip props for an equipped item's anchor, with the suffix and enchant it carries. */
export async function equippedItemTooltipProps(equippedItem: EquippedItem): Promise<Record<string, string>> {
	if (!(await hasItemTooltip(equippedItem.id))) return {};
	return itemTooltipAnchor(equippedItem.id, { randomSuffix: equippedItem.randomSuffix?.id, enchant: equippedItem.enchant?.effectId });
}
