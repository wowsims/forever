import * as InputHelpers from '@features/inputs/input_helpers';
import { Spec, UnitReference, UnitReference_Type as UnitType } from '@generated/proto/common';
import i18n from '@i18n/config';
import type { Player } from '@sim/player/player';
import { emptyUnitReference } from '@sim/proto/utils';

const assignment = (key: string) => i18n.t(`settings_tab.other.innervate_assignment.${key}`);

const isInnervateOnOther = (player: Player<Spec.SpecBalanceDruid>): boolean => {
	if ((player.getRaid()?.getTargetDummies() ?? 0) < 1) {
		return false;
	}
	const ref = player.getClassOptions().innervateTarget;
	if (!ref || ref.type === UnitType.Unknown || ref.type === UnitType.Self) {
		return false;
	}
	return !(ref.type === UnitType.Player && ref.index === player.getRaidIndex());
};

// The first empty raid slot is where NewRaid places dummy 1.
const firstEmptyRaidIndex = (player: Player<Spec.SpecBalanceDruid>): number => {
	const players = player.getRaid()?.getActivePlayers() ?? [];
	if (players.length === 0) {
		return 1;
	}
	const taken = new Set(players.map(p => p.getRaidIndex()));
	let index = 0;
	while (taken.has(index)) {
		index += 1;
	}
	return index;
};

export const InnervateAssignmentInput: InputHelpers.TypedChoiceFieldConfig<Player<Spec.SpecBalanceDruid>> = {
	id: 'innervate-assignment',
	type: 'enum',
	label: i18n.t('settings_tab.other.innervate_assignment.label'),
	labelTooltip: i18n.t('settings_tab.other.innervate_assignment.tooltip'),
	storeField: ['specOptions', 'raid:targetDummies'] as const,
	values: [
		{ name: assignment('self'), value: 0 },
		{ name: assignment('other'), value: 1 },
	],
	getValue: player => (isInnervateOnOther(player) ? 1 : 0),
	setValue: (player, value) => {
		const options = player.getClassOptions();
		if (value === 1) {
			options.innervateTarget = UnitReference.create({ type: UnitType.Player, index: firstEmptyRaidIndex(player) });
			const raid = player.getRaid();
			if (raid && raid.getTargetDummies() < 1) {
				raid.setTargetDummies(1);
			}
		} else {
			options.innervateTarget = emptyUnitReference();
		}
		player.setClassOptions(options);
	},
};
