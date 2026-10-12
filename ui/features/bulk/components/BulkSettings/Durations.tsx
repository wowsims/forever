import i18n from '@i18n/config';
import { usePlayer, useSim } from '@sim/context/SimHostContext';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { NumberField } from '@ui-kit/TextField';
import { useId } from 'react';
import { useStore } from 'zustand';

import { useBulkState } from '../../hooks/useBulkState';
import { addBulkDuration, type DurationPair, isDuplicateDuration, removeBulkDuration, updateBulkDuration } from '../../model/durations';

export interface DurationRow extends DurationPair {
	duplicate: boolean;
}

export interface DurationsViewProps {
	encounter: DurationPair;
	rows: readonly DurationRow[];
	healthFight: boolean;
	onAdd: () => void;
	onChange: (index: number, patch: Partial<DurationPair>) => void;
	onRemove: (index: number) => void;
}

const t = (key: string, options?: Record<string, unknown>) => i18n.t(`bulk_tab.settings.durations.${key}`, options);

export const DurationsView = ({ encounter, rows, healthFight, onAdd, onChange, onRemove }: DurationsViewProps) => {
	const idPrefix = useId();
	const pairFields = (id: string, rowLabel: string, pair: DurationPair, disabled: boolean, onPairChange?: (patch: Partial<DurationPair>) => void) => (
		<>
			<label className="sr-only" htmlFor={`${id}-duration`}>
				{`${rowLabel} ${t('duration')}`}
			</label>
			<NumberField
				id={`${id}-duration`}
				inputSize="sm"
				positive
				value={pair.duration}
				disabled={disabled}
				onChange={value => onPairChange?.({ duration: Math.max(1, value) })}
			/>
			<label className="sr-only" htmlFor={`${id}-variation`}>
				{`${rowLabel} ${t('variation')}`}
			</label>
			<NumberField
				id={`${id}-variation`}
				inputSize="sm"
				positive
				value={pair.durationVariation}
				disabled={disabled}
				onChange={value => onPairChange?.({ durationVariation: Math.max(0, value) })}
			/>
		</>
	);

	return (
		<div className="grid gap-2" data-testid="bulk-durations">
			<p className="mb-0 text-sm">{t('description')}</p>
			{healthFight && <p className="mb-0 text-sm text-gray-400">{t('health_fight')}</p>}
			<div className="grid grid-cols-[1fr_auto_auto_auto] items-center gap-2">
				<span />
				<span className="text-xs text-gray-400">{t('duration')}</span>
				<span className="text-xs text-gray-400">{t('variation')}</span>
				<span />
				<span>{t('encounter')}</span>
				{pairFields(`${idPrefix}-encounter`, t('encounter'), encounter, true)}
				<Icon name="lock" className="text-gray-400" />
				{rows.map((row, index) => (
					<div key={index} className="contents" data-testid="bulk-duration-row">
						<span className="grid">
							{t('extra', { index: index + 1 })}
							{row.duplicate && <span className="text-xs text-warning">{t('duplicate')}</span>}
						</span>
						{pairFields(`${idPrefix}-${index}`, t('extra', { index: index + 1 }), row, healthFight, patch => onChange(index, patch))}
						<Button
							iconOnly
							aria-label={t('remove', { index: index + 1 })}
							className="text-link-danger"
							disabled={healthFight}
							onClick={() => onRemove(index)}>
							<Icon name="xmark" />
						</Button>
					</div>
				))}
			</div>
			<Button variant="outline-primary" size="sm" disabled={healthFight} onClick={onAdd}>
				<Icon name="plus" />
				{t('add')}
			</Button>
		</div>
	);
};

export const Durations = () => {
	const player = usePlayer();
	const sim = useSim();
	const duration = useStore(sim.store, state => state.encounter.duration);
	const durationVariation = useStore(sim.store, state => state.encounter.durationVariation);
	const useHealth = useStore(sim.store, state => state.encounter.useHealth);
	const durations = useBulkState(slice => slice.durations);
	const encounter = { duration, durationVariation };

	return (
		<DurationsView
			encounter={encounter}
			rows={durations.map((row, index) => ({
				duration: row.duration,
				durationVariation: row.durationVariation,
				duplicate: !useHealth && isDuplicateDuration(encounter, durations, index),
			}))}
			healthFight={useHealth}
			onAdd={() => addBulkDuration(player, encounter)}
			onChange={(index, patch) => updateBulkDuration(player, index, patch)}
			onRemove={index => removeBulkDuration(player, index)}
		/>
	);
};
