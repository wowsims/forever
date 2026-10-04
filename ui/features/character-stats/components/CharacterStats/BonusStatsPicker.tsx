import { ConnectedNumberField } from '@features/inputs/components/ConnectedTextField';
import type { NumberFieldConfig } from '@features/inputs/configs';
import { usePlayer } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import type { UnitStat } from '@sim/proto/stats';
import { useMemo } from 'react';

export interface BonusStatsPickerProps {
	unitStat: UnitStat;
	label: string;
	onCommit: () => void;
}

/**
 * Bonus stats include pseudo stats — the speed multipliers have no root stat at all — so
 * the picker keys on the UnitStat rather than on a Stat.
 */
export const BonusStatsPicker = ({ unitStat, label, onCommit }: BonusStatsPickerProps) => {
	const player = usePlayer();

	const config = useMemo(
		(): NumberFieldConfig<Player<any>> => ({
			id: `character-bonus-stat-${unitStat.getKey()}`,
			label,
			inputClassName: 'm-0 w-32 flex-1',
			storeField: 'bonusStats',
			getValue: subject => {
				const bonusStats = subject.getBonusStats();
				return unitStat.hasRootStat() ? bonusStats.getStat(unitStat.getRootStat()) : bonusStats.getUnitStat(unitStat);
			},
			setValue: (subject, newValue) => {
				const bonusStats = subject.getBonusStats();
				subject.setBonusStats(
					unitStat.hasRootStat()
						? bonusStats.withStat(unitStat.getRootStat(), newValue)
						: unitStat.isPseudoStat()
							? bonusStats.withPseudoStat(unitStat.getPseudoStat(), newValue)
							: bonusStats.withStat(unitStat.getStat(), newValue),
				);
				onCommit();
			},
		}),
		[unitStat, label, onCommit],
	);

	return <ConnectedNumberField modObject={player} config={config} />;
};
