import { ConnectedNumberField } from '@features/inputs/components/ConnectedTextField';
import type { NumberFieldConfig } from '@features/inputs/configs';
import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import type { UnitStat } from '@sim/proto/stats';
import { ToggleGroup } from '@ui-kit/ToggleGroup';
import { useCallback, useMemo, useRef, useState } from 'react';

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
	const rootStat = unitStat.hasRootStat() ? unitStat.getRootStat() : undefined;
	const supportsPercent = unitStat.isPercentRating();
	const [inputUnit, setInputUnit] = useState<'rating' | 'percent'>(supportsPercent ? 'percent' : 'rating');
	const percentInput = supportsPercent && inputUnit === 'percent';
	const ratingSuffix = i18n.t('sidebar.character_stats.rating_suffix');
	const percentSuffix = i18n.t('sidebar.character_stats.percent_suffix');
	const percentLabel = i18n.t('sidebar.character_stats.percent_label');
	const inputLabel = percentInput ? `${label.endsWith(ratingSuffix) ? label.slice(0, -ratingSuffix.length).trimEnd() : label} (${percentSuffix})` : label;
	const toggle = useRef<HTMLDivElement>(null);
	// Safari and Firefox on macOS do not focus a clicked button, so a press on the toggle is
	// tracked by pointer; keyboard focus moving onto the toggle is caught one frame later.
	const pressedToggle = useRef(false);
	const commitInput = useCallback(() => {
		if (!supportsPercent) {
			onCommit();
			return;
		}

		requestAnimationFrame(() => {
			if (pressedToggle.current || toggle.current?.contains(document.activeElement)) return;
			onCommit();
		});
	}, [onCommit, supportsPercent]);

	const config = useMemo(
		(): NumberFieldConfig<Player<any>> => ({
			id: `character-bonus-stat-${unitStat.getKey()}`,
			label: inputLabel,
			inputClassName: 'm-0 w-32 flex-1',
			float: percentInput,
			maxDecimalDigits: percentInput ? 3 : undefined,
			storeField: 'bonusStats',
			getValue: subject => {
				const bonusStats = subject.getBonusStats();
				const value = rootStat !== undefined ? bonusStats.getStat(rootStat) : bonusStats.getUnitStat(unitStat);
				return percentInput ? unitStat.convertRatingToPercent(value)! : value;
			},
			setValue: (subject, newValue) => {
				const bonusStats = subject.getBonusStats();
				// Ratings are whole numbers in game.
				const value = percentInput ? Math.round(unitStat.convertPercentToRating(newValue)!) : newValue;
				subject.setBonusStats(rootStat !== undefined ? bonusStats.withStat(rootStat, value) : bonusStats.withUnitStat(unitStat, value));
				commitInput();
			},
		}),
		[unitStat, rootStat, inputLabel, percentInput, commitInput],
	);

	const input = <ConnectedNumberField key={inputUnit} modObject={player} config={config} />;
	if (!supportsPercent) return input;

	return (
		<div className="flex flex-col gap-2" onPointerDownCapture={event => (pressedToggle.current = !!toggle.current?.contains(event.target as Node))}>
			<div ref={toggle}>
				<ToggleGroup
					options={[
						{ value: 'rating', label: ratingSuffix },
						{ value: 'percent', label: percentLabel },
					]}
					value={inputUnit}
					onChange={setInputUnit}
					aria-label={label}
					size="xs"
					testId="bonus-stats-unit-toggle"
				/>
			</div>
			{input}
		</div>
	);
};
