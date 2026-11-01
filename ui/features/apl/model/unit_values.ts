import { UnitReference, UnitReference_Type as UnitType } from '@generated/proto/common';
import i18n from '@i18n/config';
import type { Player } from '@sim/player/player';
import { type ActionId, defaultTargetIcon, getPetIconFromName } from '@sim/proto/action_id';
import { MAX_PARTY_SIZE } from '@sim/raid/party';
import type { ResolvedIcon } from '@ui-kit/Field/types';
import type { IconName } from '@ui-kit/Icon';

import { healTargetSlots, isRaidHealing, type UNIT_SET, unitSets } from './unit_sets';

const label = (key: string) => i18n.t(`rotation_tab.apl.helpers.unit_labels.${key}`);

/** One unit as a unit field offers it: the reference, and the text, icon and class colour it shows with. */
export interface UnitValue {
	value: UnitReference | undefined;
	text?: string;
	icon?: IconName;
	iconUrl?: string;
	color?: string;
}

/** A `UnitValue` whose icon may still be an `ActionId` to resolve (a pet's). */
export interface UnitModel extends Omit<UnitValue, 'iconUrl'> {
	iconUrl?: string | ActionId;
}

/** The `UnitValue` a unit picker takes: an `ActionId` icon becomes the url it resolved to, or none yet. */
export const withIconUrl = (unit: UnitModel, resolved?: ResolvedIcon): UnitValue => ({
	...unit,
	iconUrl: typeof unit.iconUrl === 'string' ? unit.iconUrl : resolved?.iconUrl,
});

/** The icon, text and reference one unit shows as; a caller labelling many units passes the heal target slots it read once. */
export const refToValue = (ref: UnitReference | undefined, player: Player<any>, targetUI: boolean | undefined, healTargets?: Array<number>): UnitModel => {
	if (!ref || ref.type == UnitType.Unknown) {
		return { value: ref, icon: targetUI ? 'bullseye' : 'user', text: targetUI ? label('current_target') : label('self') };
	}
	switch (ref.type) {
		case UnitType.Self:
			return { value: ref, icon: 'user', text: label('self') };
		case UnitType.CurrentTarget:
			return { value: ref, icon: 'bullseye', text: label('current_target') };
		case UnitType.PreviousTarget:
			return { value: ref, icon: 'arrow-left', text: label('previous_target') };
		case UnitType.NextTarget:
			return { value: ref, icon: 'arrow-right', text: label('next_target') };
		case UnitType.RandomFriendlyTarget:
			return { value: ref, icon: 'user', text: label('random_friendly_target') };
		case UnitType.Player: {
			const other = player.sim.raid.getPlayer(ref.index);
			if (other) return { value: ref, iconUrl: other.getSpecIcon(), text: `${label('player')} ${ref.index + 1}` };
			const healTargetIndex = (healTargets ?? healTargetSlots(player)).indexOf(ref.index);
			// Numbered across the whole raid, in the order the results filter numbers the same units.
			return healTargetIndex >= 0 ? { value: ref, icon: 'user', text: `${label('heal_target')} ${healTargetIndex + 1}` } : { value: ref };
		}
		case UnitType.Target: {
			const targetMetadata = player.sim.encounter.targetsMetadata.asList()[ref.index];
			return targetMetadata ? { value: ref, iconUrl: defaultTargetIcon, text: `${label('target')} ${ref.index + 1}` } : { value: ref };
		}
		case UnitType.Pet: {
			const petMetadata = player.sim.getUnitMetadata(ref, player, UnitReference.create({ type: UnitType.Self }));
			let text = `${label('pet')} ${ref.index + 1}`;
			let iconUrl: string | ActionId | undefined;
			const petName = petMetadata?.getName();
			if (petName) {
				text = petName.substring(petName.indexOf(' - ') + ' - '.length);
				iconUrl = getPetIconFromName(text);
			}
			return { value: ref, icon: iconUrl ? undefined : 'paw', iconUrl, text };
		}
		default:
			return { value: ref };
	}
};

export interface UnitOptionModel {
	unit: UnitModel;
	/** Pets are filed under their owner, and healing targets under the heal-target submenu. */
	submenu?: Array<UnitModel | string>;
}

/** The unit set's options, each with the text, icon and colour it displays with; a pet's icon may still be an `ActionId` to resolve. */
export const unitOptionModels = (unitSet: UNIT_SET, player: Player<any>): Array<UnitOptionModel> => {
	const set = unitSets[unitSet];
	const includesHealingTargets = unitSet !== 'self_and_pets';
	const healingTargets = includesHealingTargets ? healTargetSlots(player) : [];
	const raidHealing = healingTargets.length > 0 && isRaidHealing(player);
	const models = set.getUnits(player).map(ref => {
		const unit = refToValue(ref, player, set.targetUI, healingTargets);
		if (ref?.type === UnitType.Player && healingTargets.includes(ref.index)) {
			if (!raidHealing) return { unit, submenu: ['heal_target'] };
			const group = Math.floor(ref.index / MAX_PARTY_SIZE) + 1;
			return { unit, submenu: ['heal_target', `group_${group}`] };
		}
		if (ref?.type != UnitType.Pet) return { unit };
		return { unit, submenu: [refToValue(set.targetUI ? ref.owner! : undefined, player, set.targetUI, healingTargets)] };
	});
	if (raidHealing) {
		models.push({
			unit: refToValue(UnitReference.create({ type: UnitType.RandomFriendlyTarget }), player, set.targetUI, healingTargets),
			submenu: ['heal_target'],
		});
	}
	return models;
};
