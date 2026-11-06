import type { ChoiceFieldConfig, NumberFieldConfig } from '@features/inputs/configs';
import type { Target } from '@generated/proto/common';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { ChoiceField } from '@ui-kit/ChoiceField';
import { FieldGrid } from '@ui-kit/FieldGrid';
import { NumberField } from '@ui-kit/TextField';

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
} from './utils/configs';

export interface PlannerAddStatsProps {
	/** A wave's add or an encounter target. */
	target: Target;
	/** Edits the target: in the draft for an add, in the store for an encounter target. */
	onModify: (edit: (target: Target) => void) => void;
	/** Starts the field ids, so two waves never share one. */
	idPrefix: string;
	/** The NPC and preset target fields of an encounter target, shown first on a row each; an add's template field picks its preset instead. */
	identity?: ReadonlyArray<ChoiceFieldConfig<null>>;
	/** Shows the fields only a threat sim reads: attack power, block value, the tank, the swing, the melee toggles and the spell school. */
	threatMetrics?: boolean;
}

/**
 * The fields of a target: who it is, its defences and how it attacks, for a wave's add and for an encounter target
 * alike. An add shows the threat fields with threat metrics off too: the tank stays, since the sim reads the
 * template's tank index to tell who an add attacks.
 */
export const PlannerAddStats = ({ target, onModify, idPrefix, identity = [], threatMetrics = true }: PlannerAddStatsProps) => {
	const context: TargetFieldContext = { idPrefix, getTarget: () => target, modifyTarget: onModify };

	const choice = (config: ChoiceFieldConfig<null>, className?: string) => (
		<ChoiceField
			key={config.id}
			id={config.id}
			className={className}
			label={config.label}
			labelTooltip={config.labelTooltip}
			options={config.values.map(entry => ({ value: entry.value, label: entry.name, tooltip: entry.tooltip }))}
			value={config.getValue(null)}
			onChange={value => config.setValue(null, value)}
		/>
	);

	return (
		<FieldGrid>
			{identity.map(config => choice(config, 'col-span-full'))}
			{choice(mobTypeConfig(context))}
			{creatureFamiliesOf(target.mobType).length > 0 && choice(creatureFamilyConfig(context, target.mobType))}
			{choice(levelConfig(context))}
			{ALL_TARGET_STATS.filter(entry => threatMetrics || !entry.threatMetrics)
				.map((entry): NumberFieldConfig<null> => statConfig(context, entry.stat, entry.tooltip))
				.map(config => (
					<NumberField
						key={config.id}
						id={config.id}
						label={config.label}
						labelTooltip={config.labelTooltip}
						value={config.getValue(null)}
						onChange={value => config.setValue(null, value)}
					/>
				))}
			{(threatMetrics ? numberConfigs(context) : []).map(config => (
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
			{[...(threatMetrics ? booleanConfigs(context) : []), bleedImmuneConfig(context)].map(config => (
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
			{threatMetrics && choice(tankIndexConfig(context))}
			{threatMetrics && choice(spellSchoolConfig(context))}
		</FieldGrid>
	);
};
