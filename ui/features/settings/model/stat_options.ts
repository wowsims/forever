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

export interface MultiIconPickerStatOption extends PickerStatOption<MultiIconPickerConfig<Player<any>>> {}

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

// An option is offered when it carries no stat, a stat in `shown`, or the spec includes its config.
// One whose stats are all hidden comes back as a copy offered only while picked, so the rows keep
// their places at every filter setting. Exclusions drop the option outright.
export function relevantStatOptions<OptionsType extends PickerStatOptions>(
	options: StatOptions<OptionsType>,
	config: Pick<SpecStats, 'includeBuffDebuffInputs' | 'excludeBuffDebuffInputs'>,
	shown: ReadonlySet<Stat>,
): StatOptions<OptionsType> {
	const listed = (list: ReadonlyArray<unknown>, option: OptionsType) => list.includes(option.config) || option.stats.some(stat => list.includes(stat));

	return options
		.filter(option => !listed(config.excludeBuffDebuffInputs, option))
		.map(option =>
			option.stats.length === 0 || option.stats.some(stat => shown.has(stat)) || config.includeBuffDebuffInputs.includes(option.config)
				? option
				: keptWhilePicked(option),
		);
}

// A class never buffs itself with its own buff, so on that class's settings tab the row reads
// "(External)": an outside caster is the only source. Every other option comes back as the same
// object, so include / exclude lists that name a config keep matching it by reference.
export function applyOwnerClassLabels<OptionsType extends RenderableStatOptions>(options: ReadonlyArray<OptionsType>, player: Player<any>): OptionsType[] {
	const playerClass = player.getClass();
	return options.map(option => {
		const label = option.config.label;
		if (option.ownerClass !== playerClass || typeof label !== 'string') return option;
		return { ...option, config: { ...option.config, label: i18n.t('settings_tab.external_label', { label }) } };
	});
}

const describeInput = (config: RenderableStatOptions['config']): string =>
	config.label ?? ('actionId' in config ? `the input for spell ${config.actionId.spellId}` : 'an unlabelled input');

// A registry lists its rows in display order: a prebuilt option's config stands for that option,
// keeping the stat tags and owner class it was built with, and a literal row is written out in
// place. Both directions throw: a config with no prebuilt option, and a prebuilt option no row
// names. A regenerated registry therefore fails loudly instead of losing a row off the tab.
export function inDisplayOrder(
	prebuilt: ReadonlyArray<RenderableStatOptions>,
	rows: ReadonlyArray<RenderableStatOptions | RenderableStatOptions['config']>,
): RenderableStatOptions[] {
	const composed = rows.map(row => {
		if ('config' in row) return row;
		const option = prebuilt.find(candidate => candidate.config === row);
		if (!option) throw new Error(`the display order names ${describeInput(row)}, which no prebuilt row carries`);
		return option;
	});

	const placed = new Set(composed.map(option => option.config));
	const dropped = prebuilt.filter(option => !placed.has(option.config));
	if (dropped.length > 0) {
		throw new Error(`the display order leaves out ${dropped.map(option => describeInput(option.config)).join(', ')}`);
	}

	return composed;
}
