import type { BooleanPickerConfig, ChoiceFieldConfig, NumberFieldConfig } from '@features/inputs/configs';
import type { Target } from '@generated/proto/common';
import { translateMobType, translateSpellSchool } from '@i18n/localization';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { ChoiceField } from '@ui-kit/ChoiceField';
import { FieldGrid } from '@ui-kit/FieldGrid';
import { NumberField } from '@ui-kit/TextField';

import { plannerText } from '../EncounterPlanner/utils/configs';
import { formatSeconds } from '../EncounterPlanner/utils/tips';
import { type PlannerSection, PlannerSections, type PlannerSectionState } from '../PlannerSections';
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
	type TargetSection,
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
	/** The sections the player opened or closed, which the planner keeps from one target to the next. */
	sections: PlannerSectionState;
}

/**
 * The fields of a target: who it is, its defences and how it attacks, for a wave's add and for an encounter target
 * alike. An add shows the threat fields with threat metrics off too: the tank stays, since the sim reads the
 * template's tank index to tell who an add attacks.
 */
export const PlannerAddStats = ({ target, onModify, idPrefix, identity = [], threatMetrics = true, sections }: PlannerAddStatsProps) => {
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

	const number = (config: NumberFieldConfig<null>) => (
		<NumberField
			key={config.id}
			id={config.id}
			label={config.label}
			labelTooltip={config.labelTooltip}
			float={config.float}
			value={config.getValue(null)}
			onChange={value => config.setValue(null, value)}
		/>
	);

	const toggle = (config: BooleanPickerConfig<null>) => (
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
	);

	const stats = (section: TargetSection) => ALL_TARGET_STATS.filter(entry => entry.section === section);
	const statFields = (section: TargetSection) => stats(section).map(entry => number(statConfig(context, entry.stat, entry.tooltip)));
	const summary = (key: string, options?: Record<string, string | number>) => plannerText(`panel.summary.${key}`, options);
	const resisted = stats('resistances').filter(entry => (target.stats[entry.stat] ?? 0) !== 0);

	const list: Array<PlannerSection> = [
		{
			id: 'basics',
			title: plannerText('panel.sections.basics'),
			summary: plannerText('tip.add_group', { level: target.level, type: translateMobType(target.mobType) }),
			children: (
				<FieldGrid>
					{choice(mobTypeConfig(context))}
					{creatureFamiliesOf(target.mobType).length > 0 && choice(creatureFamilyConfig(context, target.mobType))}
					{choice(levelConfig(context))}
					{statFields('basics')}
					{toggle(bleedImmuneConfig(context))}
				</FieldGrid>
			),
		},
		{
			id: 'resistances',
			title: plannerText('panel.sections.resistances'),
			summary: resisted.length
				? resisted.map(entry => summary('resistance', { school: translateSpellSchool(entry.school!), value: target.stats[entry.stat] })).join(', ')
				: summary('all_zero'),
			startsClosed: resisted.length === 0,
			children: <FieldGrid>{statFields('resistances')}</FieldGrid>,
		},
	];
	if (threatMetrics)
		list.push({
			id: 'melee',
			title: plannerText('panel.sections.melee'),
			summary: summary('melee', { damage: target.minBaseDamage, speed: formatSeconds(target.swingSpeed) }),
			children: (
				<FieldGrid>
					{statFields('melee')}
					{numberConfigs(context).map(number)}
					{booleanConfigs(context).map(toggle)}
					{choice(tankIndexConfig(context))}
					{choice(spellSchoolConfig(context))}
				</FieldGrid>
			),
		});

	return (
		<>
			{identity.length > 0 && <FieldGrid>{identity.map(config => choice(config, 'col-span-full'))}</FieldGrid>}
			<PlannerSections sections={list} state={sections} testId={`${idPrefix}-sections`} />
		</>
	);
};
