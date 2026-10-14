import i18n from '@i18n/config';
import type { PrecisionSettings } from '@sim/settings/precision';
import type { PrecisionMode } from '@sim/state/sim_store';
import { NumberField } from '@ui-kit/TextField';
import { ToggleGroup } from '@ui-kit/ToggleGroup';

export interface PrecisionOptionsProps {
	value: PrecisionSettings;
	onChange: (patch: Partial<PrecisionSettings>) => void;
	/** The sim's iteration setting, shown while `value.iterations` is 0. */
	defaultIterations: number;
	idPrefix: string;
	targetLabel: string;
	targetTooltip: string;
	iterationsTooltip: string;
	maxMultiplierTooltip: string;
	/** The target's decimals. */
	maxDecimalDigits: number;
	disabled?: boolean;
}

const t = (key: string) => i18n.t(`common.precision.${key}`);

export const PrecisionOptions = ({
	value,
	onChange,
	defaultIterations,
	idPrefix,
	targetLabel,
	targetTooltip,
	iterationsTooltip,
	maxMultiplierTooltip,
	maxDecimalDigits,
	disabled,
}: PrecisionOptionsProps) => (
	<div className="flex flex-col gap-2">
		<div className="flex flex-col items-start gap-1">
			<span className="text-sm font-semibold">{t('label')}</span>
			<ToggleGroup<PrecisionMode>
				size="sm"
				aria-label={t('label')}
				options={[
					{ value: 'iterations', label: t('mode_iterations') },
					{ value: 'target', label: t('mode_target') },
				]}
				value={value.precisionMode}
				onChange={precisionMode => onChange({ precisionMode })}
				disabled={disabled}
			/>
		</div>
		{value.precisionMode === 'target' ? (
			<div className="grid grid-cols-2 gap-4">
				<NumberField
					id={`${idPrefix}-target`}
					label={targetLabel}
					labelTooltip={targetTooltip}
					float
					positive
					maxDecimalDigits={maxDecimalDigits}
					disabled={disabled}
					value={value.targetCi}
					onChange={targetCi => onChange({ targetCi })}
				/>
				<NumberField
					id={`${idPrefix}-max-multiplier`}
					label={t('max_multiplier')}
					labelTooltip={maxMultiplierTooltip}
					positive
					showZeroes={false}
					placeholder={t('max_multiplier_none')}
					disabled={disabled}
					value={value.maxIterationMultiplier}
					onChange={maxIterationMultiplier => onChange({ maxIterationMultiplier })}
				/>
			</div>
		) : (
			<NumberField
				id={`${idPrefix}-iterations`}
				label={t('iterations')}
				labelTooltip={iterationsTooltip}
				positive
				showZeroes={false}
				placeholder={String(defaultIterations)}
				disabled={disabled}
				value={value.iterations}
				onChange={iterations => onChange({ iterations })}
			/>
		)}
	</div>
);
