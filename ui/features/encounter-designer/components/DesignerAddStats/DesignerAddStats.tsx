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
import type { NumberFieldConfig } from '@features/inputs/configs';
import { Target } from '@generated/proto/common';
import type { StoreSubscribe } from '@sim/state/subscriptions';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { ChoiceField } from '@ui-kit/ChoiceField';
import { FieldGrid } from '@ui-kit/FieldGrid';
import { NumberField } from '@ui-kit/TextField';
import { useMemo } from 'react';

export interface DesignerAddStatsProps {
	/** The target of a wave's add; a wave that has none shows the defaults. */
	target: Target | undefined;
	/** Edits the add's target in the draft. */
	onModify: (edit: (target: Target) => void) => void;
	/** Starts the field ids, so two waves never share one. */
	idPrefix: string;
}

// The preset and AI fields stay off: an add's template field picks the preset, and the sim attaches no AI to an add. The
// tank stays, since the sim reads the template's tank index to tell who an add attacks.
const noSubscribe = (): StoreSubscribe => () => () => {};

/** The fields of an encounter target that describe an add: who it is, its defences and how it attacks. The configs are the targets picker's own. */
export const DesignerAddStats = ({ target, onModify, idPrefix }: DesignerAddStatsProps) => {
	const current = target ?? Target.create();
	const context = useMemo(
		(): TargetFieldContext => ({ idPrefix, getTarget: () => current, modifyTarget: onModify, storeSubscribe: () => noSubscribe() }),
		[idPrefix, current, onModify],
	);
	const level = levelConfig(context);
	const mobType = mobTypeConfig(context);
	const creatureFamily = creatureFamilyConfig(context, current.mobType);
	const stats: Array<NumberFieldConfig<null>> = ALL_TARGET_STATS.map(entry => statConfig(context, entry.stat, entry.tooltip));
	const numbers = numberConfigs(context);
	const booleans = [...booleanConfigs(context), bleedImmuneConfig(context)];
	const spellSchool = spellSchoolConfig(context);
	const tankIndex = tankIndexConfig(context);

	const choice = (config: typeof level) => (
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
			{choice(mobType)}
			{creatureFamiliesOf(current.mobType).length > 0 && choice(creatureFamily)}
			{choice(level)}
			{stats.map(config => (
				<NumberField
					key={config.id}
					id={config.id}
					label={config.label}
					labelTooltip={config.labelTooltip}
					value={config.getValue(null)}
					onChange={value => config.setValue(null, value)}
				/>
			))}
			{numbers.map(config => (
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
			{booleans.map(config => (
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
			{choice(tankIndex)}
			{choice(spellSchool)}
		</FieldGrid>
	);
};
