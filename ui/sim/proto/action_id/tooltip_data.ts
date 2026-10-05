import type { Player } from '../../player/player';
import type { EquippedItem } from '../equipped_item';
import type { WowheadTooltipItemParams, WowheadTooltipSpellParams } from '../wowhead';
import { ActionId } from './index';

export function actionIdWowheadTooltipData(
	actionId: ActionId,
	params?: Omit<WowheadTooltipItemParams, 'itemId'> | Omit<WowheadTooltipSpellParams, 'spellId'>,
): Promise<string> {
	return actionId.itemId
		? ActionId.makeItemTooltipData(actionId.itemId, params)
		: ActionId.makeSpellTooltipData(actionId.spellIdTooltipOverride || actionId.spellId, params);
}

export function equippedItemWowheadTooltipData(player: Player<any>, equippedItem: EquippedItem): Promise<string> {
	const enchantIds = [equippedItem.enchant?.effectId].filter((id): id is number => id !== undefined);
	return actionIdWowheadTooltipData(equippedItem.asActionId(), {
		itemLevel: Number(equippedItem.ilvl),
		enchantIds: enchantIds,
		randomEnchantmentId: equippedItem.randomSuffix?.id,
		setPieceIds: player
			.getGear()
			.asArray()
			.filter(ei => ei != null)
			.map(ei => ei!.item.id),
	});
}
