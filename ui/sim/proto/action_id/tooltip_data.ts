import type { EquippedItem } from '../equipped_item';
import type { Gear } from '../gear';
import { hasSpellTooltip, spellTooltipAnchor } from '../tooltips';
import type { WowheadTooltipItemParams, WowheadTooltipSpellParams } from '../wowhead';
import { ActionId } from './index';

/** The tooltip props for an action's anchor: a spell with a tooltip of its own shows that, anything else wowhead's. */
export async function actionIdTooltipProps(actionId: ActionId, options: { useBuffAura?: boolean } = {}): Promise<Record<string, string>> {
	const spellId = actionId.itemId ? 0 : actionId.spellIdTooltipOverride || actionId.spellId;
	if (spellId && (await hasSpellTooltip(spellId))) return spellTooltipAnchor(spellId, { buff: options.useBuffAura });
	return { 'data-wowhead': await actionIdWowheadTooltipData(actionId, options) };
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
