import { inputPicked, valuePicked } from '@features/inputs/components/ConnectedIconEnumPicker/utils';
import type { IconEnumPickerConfig, IconPickerConfig, MultiIconPickerConfig } from '@features/inputs/configs';
import type { StatFilterScope } from '@features/stat-filters/model/stat_filters';
import { Class, Stat } from '@generated/proto/common';
import i18n from '@i18n/config';
import { Player } from '@sim/player/player';
import { UnitStat } from '@sim/proto/stats';
import type { IndividualSimUIConfig } from '@sim/spec_config';

export interface StatOption {
	stats: Array<Stat>;
	// The class that casts this buff, where one class owns it. Set by the generated buff rows.
	ownerClass?: Class;
}

export interface PickerStatOption<ConfigType> extends StatOption {
	config: ConfigType;
}

export interface IconPickerStatOption extends PickerStatOption<IconPickerConfig<Player<any>, any>> {}

export interface MultiIconPickerStatOption extends PickerStatOption<MultiIconPickerConfig<Player<any>>> {
	// For a group of buffs that do not stack, the rows its inputs come from, in the same order: each keeps its own stat
	// tags and owner class.
	members?: IconPickerStatOption[];
}

export interface IconEnumPickerStatOption extends PickerStatOption<IconEnumPickerConfig<Player<any>, any>> {}

export type PickerStatOptions = IconPickerStatOption | MultiIconPickerStatOption | IconEnumPickerStatOption;
export type RenderableStatOptions = IconPickerStatOption | MultiIconPickerStatOption | IconEnumPickerStatOption;
export type StatOptions<Options extends PickerStatOptions> = Array<Options>;

export const isMultiIconOption = (option: PickerStatOptions): option is MultiIconPickerStatOption => 'inputs' in option.config;
export const isIconEnumOption = (option: PickerStatOptions): option is IconEnumPickerStatOption => 'values' in option.config;

type SpecStats = Pick<IndividualSimUIConfig<any>, 'displayStats' | 'epStats' | 'includeBuffDebuffInputs' | 'excludeBuffDebuffInputs'>;

// A spec's includeBuffDebuffInputs / excludeBuffDebuffInputs list holds stats and input configs. An
// included stat is one of the filter's defaults; an excluded stat drops every option tagged with it.
// A listed config forces that one option on or off, so a spec can drop a single buff that shares its
// stat tag with buffs it wants to keep, e.g. Mana Tide but not Mana Spring.
//
// The defaults are the displayed stats (a pseudo-stat by its root rating), the EP stats and the stats the spec
// includes. The universe is every stat tagged on the options, once and in stat order, less the excluded stats: a
// toggle for one of those could never take effect.
export const statOptionScope = (options: ReadonlyArray<PickerStatOptions>, config: SpecStats): StatFilterScope => {
	const displayed = config.displayStats.filter(us => us.hasRootStat()).map(us => us.getRootStat());
	const included = config.includeBuffDebuffInputs.filter((entry): entry is Stat => typeof entry === 'number');
	const tagged = new Set(options.flatMap(option => option.stats));
	return {
		defaults: [...new Set([...displayed, ...config.epStats, ...included])].map(stat => UnitStat.fromStat(stat)),
		universe: [...tagged]
			.filter(stat => !config.excludeBuffDebuffInputs.includes(stat))
			.sort((a, b) => a - b)
			.map(stat => UnitStat.fromStat(stat)),
	};
};

// Buffs the sim lets only one of apply, offered as one multi icon under a label for what they give. A registry names
// the group by its members' configs; inDisplayOrder resolves them to the prebuilt rows.
export interface StatOptionGroup {
	groupLabel: string;
	configs: ReadonlyArray<IconPickerStatOption['config']>;
	// Shown on the group's icon, for a group whose inputs the label alone does not explain.
	tooltip?: string;
}

export const statOptionGroup = (label: string, configs: ReadonlyArray<IconPickerStatOption['config']>, tooltip?: string): StatOptionGroup => ({
	groupLabel: label,
	configs,
	tooltip,
});

const isGroupRow = (row: object): row is StatOptionGroup => 'groupLabel' in row;

// Offered while any member is, tagged with every stat a member carries.
const groupOf = (label: string, members: IconPickerStatOption[], tooltip?: string): MultiIconPickerStatOption => {
	const inputs = members.map(member => member.config);
	return {
		config: { label, tooltip, inputs, showWhen: player => inputs.some(input => !input.showWhen || input.showWhen(player)) },
		stats: [...new Set(members.flatMap(member => member.stats))],
		members,
	};
};

const keptWhile = <Config extends PickerStatOptions['config']>(config: Config): Config => {
	const showWhen = config.showWhen as ((player: Player<any>) => boolean) | undefined;
	return {
		...config,
		showWhen: (player: Player<any>) => (!showWhen || showWhen(player)) && inputPicked(config, player),
		keptBy: (_player: Player<any>, value: unknown) => valuePicked(config, value),
	};
};

// A kept multi icon offers only its picked inputs, each marked kept: its inputs share its hidden stats.
const keptWhilePicked = <OptionsType extends PickerStatOptions>(option: OptionsType): OptionsType => {
	const config = keptWhile(option.config);
	if (!isMultiIconOption(option)) return { ...option, config };
	return { ...option, config: { ...config, inputs: option.config.inputs.map(input => keptWhile(input)) } };
};

// An option is offered when it carries no stat, a stat in `shown`, or the spec includes its config (for a group, a member's).
// One whose stats are all hidden comes back as a copy offered only while picked, so the rows keep
// their places at every filter setting. Exclusions drop the option outright.
export function relevantStatOptions<OptionsType extends PickerStatOptions>(
	options: StatOptions<OptionsType>,
	config: Pick<SpecStats, 'includeBuffDebuffInputs' | 'excludeBuffDebuffInputs'>,
	shown: ReadonlySet<Stat>,
): StatOptions<OptionsType> {
	const listed = (list: ReadonlyArray<unknown>, option: PickerStatOptions) => list.includes(option.config) || option.stats.some(stat => list.includes(stat));

	// A group loses the members the spec excludes, and is gone with the last of them.
	const withoutExcluded = (option: OptionsType): OptionsType | undefined => {
		if (!isMultiIconOption(option) || !option.members) return listed(config.excludeBuffDebuffInputs, option) ? undefined : option;
		if (config.excludeBuffDebuffInputs.includes(option.config)) return undefined;
		const members = option.members.filter(member => !listed(config.excludeBuffDebuffInputs, member));
		if (members.length === 0) return undefined;
		return members.length === option.members.length ? option : (groupOf(option.config.label ?? '', members) as OptionsType);
	};
	const included = (option: OptionsType) =>
		config.includeBuffDebuffInputs.includes(option.config) ||
		(isMultiIconOption(option) && !!option.members?.some(member => config.includeBuffDebuffInputs.includes(member.config)));

	return options
		.map(withoutExcluded)
		.filter(option => option !== undefined)
		.map(option => (option.stats.length === 0 || option.stats.some(stat => shown.has(stat)) || included(option) ? option : keptWhilePicked(option)));
}

// A class never buffs itself with its own buff, so on that class's settings tab the row reads
// "(External)": an outside caster is the only source. Every other option comes back as the same
// object, so include / exclude lists that name a config keep matching it by reference.
export function applyOwnerClassLabels<OptionsType extends RenderableStatOptions>(options: ReadonlyArray<OptionsType>, player: Player<any>): OptionsType[] {
	const playerClass = player.getClass();
	const external = <Config extends { label?: string }>(config: Config, ownerClass: Class | undefined): Config => {
		const label = config.label;
		if (ownerClass !== playerClass || typeof label !== 'string') return config;
		return { ...config, label: i18n.t('settings_tab.external_label', { label }) };
	};
	return options.map(option => {
		// A group keeps its own label and marks its members: its inputs line up with them by index.
		if (isMultiIconOption(option) && option.members) {
			const members = option.members;
			const inputs = option.config.inputs.map((input, index) => external(input, members[index]?.ownerClass));
			return inputs.every((input, index) => input === option.config.inputs[index]) ? option : { ...option, config: { ...option.config, inputs } };
		}
		const config = external(option.config, option.ownerClass);
		return config === option.config ? option : { ...option, config };
	});
}

const describeInput = (config: RenderableStatOptions['config']): string =>
	config.label ?? ('actionId' in config ? `the input for spell ${config.actionId.spellId}` : 'an unlabelled input');

// A registry lists its rows in display order: a prebuilt option's config stands for that option,
// keeping the stat tags and owner class it was built with, a group stands for the prebuilt options
// its configs name, and a literal row is written out in place. Both directions throw: a config with
// no prebuilt option, and a prebuilt option no row names. A regenerated registry therefore fails
// loudly instead of losing a row off the tab.
export function inDisplayOrder(
	prebuilt: ReadonlyArray<RenderableStatOptions>,
	rows: ReadonlyArray<RenderableStatOptions | RenderableStatOptions['config'] | StatOptionGroup>,
): RenderableStatOptions[] {
	const resolve = (config: RenderableStatOptions['config']): RenderableStatOptions => {
		const option = prebuilt.find(candidate => candidate.config === config);
		if (!option) throw new Error(`the display order names ${describeInput(config)}, which no prebuilt row carries`);
		return option;
	};
	const member = (config: IconPickerStatOption['config']): IconPickerStatOption => {
		const option = resolve(config);
		if (isMultiIconOption(option) || isIconEnumOption(option)) throw new Error(`a group takes icon rows only, and ${describeInput(config)} is not one`);
		return option;
	};

	const composed = rows.map(row => {
		if (isGroupRow(row)) return groupOf(row.groupLabel, row.configs.map(member), row.tooltip);
		if ('config' in row) return row;
		return resolve(row);
	});

	const placed = new Set(composed.flatMap(option => [option.config, ...((isMultiIconOption(option) && option.members?.map(m => m.config)) || [])]));
	const dropped = prebuilt.filter(option => !placed.has(option.config));
	if (dropped.length > 0) {
		throw new Error(`the display order leaves out ${dropped.map(option => describeInput(option.config)).join(', ')}`);
	}

	return composed;
}
