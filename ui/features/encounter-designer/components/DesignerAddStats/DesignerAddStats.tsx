import {
	ALL_TARGET_STATS,
	bleedImmuneConfig,
	booleanConfigs,
	creatureFamiliesOf,
	creatureFamilyConfig,
	levelConfig,
	mobTypeConfig,
	numberConfigs,
	spellSchoolConfig,
	statConfig,
	tankIndexConfig,
	type TargetFieldContext,
} from '@features/encounter/components/TargetsPicker/utils/configs';
import type { ChoiceFieldConfig, NumberFieldConfig } from '@features/inputs/configs';
import type { Target } from '@generated/proto/common';
import type { StoreSubscribe } from '@sim/state/subscriptions';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { ChoiceField } from '@ui-kit/ChoiceField';
import { FieldGrid } from '@ui-kit/FieldGrid';
import { NumberField } from '@ui-kit/TextField';

export interface DesignerAddStatsProps {
	/** The target of a wave's add. */
	target: Target;
	/** Edits the add's target in the draft. */
	onModify: (edit: (target: Target) => void) => void;
	/** Starts the field ids, so two waves never share one. */
	idPrefix: string;
}

const noSubscribe = (): StoreSubscribe => () => () => {};

/**
 * The fields of an encounter target that describe an add: who it is, its defences and how it attacks. The configs are the targets picker's own. The
 * preset and AI fields stay off: an add's template field picks the preset, and the sim attaches no AI to an add. The tank stays, since the sim reads the
 * template's tank index to tell who an add attacks.
 */
export const DesignerAddStats = ({ target, onModify, idPrefix }: DesignerAddStatsProps) => {
	const context: TargetFieldContext = { idPrefix, getTarget: () => target, modifyTarget: onModify, storeSubscribe: () => noSubscribe() };

	const choice = (config: ChoiceFieldConfig<null>) => (
		<ChoiceField
			key={config.id}
			id={config.id}
			label={config.label}
			labelTooltip={config.labelTooltip}
			options={config.values.map(entry => ({ value: entry.value, label: entry.name, tooltip: entry.tooltip }))}
			value={config.getValue(null)}
			onChange={value => config.setValue(null, value)}
		/>
	);

	return (
		<FieldGrid>
			{choice(mobTypeConfig(context))}
			{creatureFamiliesOf(target.mobType).length > 0 && choice(creatureFamilyConfig(context, target.mobType))}
			{choice(levelConfig(context))}
			{ALL_TARGET_STATS.map((entry): NumberFieldConfig<null> => statConfig(context, entry.stat, entry.tooltip)).map(config => (
				<NumberField
					key={config.id}
					id={config.id}
					label={config.label}
					labelTooltip={config.labelTooltip}
					value={config.getValue(null)}
					onChange={value => config.setValue(null, value)}
				/>
			))}
			{numberConfigs(context).map(config => (
				<NumberField
					key={config.id}
					id={config.id}
					label={config.label}
					labelTooltip={config.labelTooltip}
					float={config.float}
					value={config.getValue(null)}
					onChange={value => config.setValue(null, value)}
				/>
			))}
			{[...booleanConfigs(context), bleedImmuneConfig(context)].map(config => (
				<BooleanPicker
					key={config.id}
					id={config.id}
					label={config.label}
					labelTooltip={config.labelTooltip}
					layout="inline"
					reverse
					disabled={config.enableWhen?.(null) === false}
					value={config.getValue(null)}
					onChange={value => config.setValue(null, value)}
				/>
			))}
			{choice(tankIndexConfig(context))}
			{choice(spellSchoolConfig(context))}
		</FieldGrid>
	);
};
