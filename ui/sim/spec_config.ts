import type * as IconInputs from '@features/inputs/icon_inputs';
import type * as InputHelpers from '@features/inputs/input_helpers';
import type { ReforgeOptimizerOptions } from '@features/reforge/model/reforge_optimizer';
import { APLRotation, APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { ConsumesSpec, EquipmentSpec, HealingModel, ItemSlot, ItemSwap, Profession, PseudoStat, Race, Spec, Stat } from '@generated/proto/common';
import { SavedTalents } from '@generated/proto/ui';
import type { SavedDataConfig } from '@ui-kit/SavedDataPanel/types';
import type { ReactNode } from 'react';

import { Player, PlayerConfig, registerSpecConfig as registerPlayerConfig } from './player/player';
import type { WeaponSkillGroup } from './player/weapon_skill_groups';
import type { PresetBuild, PresetEncounter, PresetEpWeights, PresetGear, PresetItemSwap, PresetRotation, PresetSettings } from './presets/types';
import type { SpecOptions, SpecRotation } from './proto/spec_types';
import type { StatMods, StatWrites } from './proto/stats';
import { StatCap, Stats, UnitStat } from './proto/stats';
import type { Sim } from './sim';
import type { IndividualSimHost, SimWarning } from './sim_host';
import type { StoreSubscribe } from './state/subscriptions';

/** Per-stat tooltip content a spec contributes to the reforge panel. */
export type StatTooltipContent = { [key in Stat]?: () => ReactNode };

export interface EncounterPickerConfig {
	showExecuteProportion: boolean;
}

export type InputConfig<ModObject> =
	| InputHelpers.TypedBooleanPickerConfig<ModObject>
	| InputHelpers.TypedNumberFieldConfig<ModObject>
	| InputHelpers.TypedChoiceFieldConfig<ModObject>;

export interface InputSection {
	tooltip?: string;
	inputs: Array<InputConfig<Player<any>>>;
}

// An extra settings-tab section, declared as data. Core renders it (see
// `SettingsTab.buildCustomSettingsSections`) with the same ContentBlock and
// picker construction path the standard sections use — a spec never builds DOM.
export interface CustomSection<SpecType extends Spec> {
	// Stable identifier. Doubles as the ContentBlock css class when `className`
	// is not given.
	id: string;
	title: string;
	tooltip?: string;
	// Body text under the section title, for a section whose explanation is too long to hide
	// behind the title's tooltip.
	description?: string;
	// Css class for the section's ContentBlock; this is what stylesheets hook on.
	className?: string;
	// Hides the whole section when this evaluates to false, mirroring `showWhen`
	// on an individual input.
	when?: (player: Player<SpecType>) => boolean;
	// Icon pickers, rendered into a group container above `inputs`.
	iconInputs?: Array<IconInputs.IconInputConfig<Player<SpecType>, any>>;
	// Extra css class for that group container; the layout hook for the icon row.
	iconGroupClassName?: string;
	// Non-icon pickers, same shape as `otherInputs.inputs`.
	inputs?: Array<InputConfig<Player<any>>>;
}

export interface OtherDefaults {
	profession1?: Profession;
	profession2?: Profession;
	distanceFromTarget?: number;
	channelClipDelay?: number;
	reactionTime?: number;
	iterationCount?: number;
	race?: Race;
	// Applied by applyIndividualDefaults after the encounter defaults.
	healingModel?: HealingModel;
}

export interface IndividualSimUIConfig<SpecType extends Spec> extends PlayerConfig<SpecType> {
	// Override for required talent rows. If not specified, defaults to requiring all rows [0, 1, 2, 3, 4, 5]
	requiredTalentRows?: number[];
	// Additional css class to add to the root element.
	className: string;
	// Used to generate schemed components. E.g. 'shaman', 'druid', 'raid'
	cssScheme: string;

	knownIssues?: Array<ReactNode>;
	warnings?: Array<(simUI: IndividualSimHost<SpecType>) => SimWarning>;
	consumableStats?: Array<Stat>;
	epStats: Array<Stat>;
	epPseudoStats?: Array<PseudoStat>;
	epReferenceStat: Stat;
	// Reference stat for tank specs' EP; falls back to StatArmor when unset.
	tankRefStat?: Stat;
	// Per-spec override for the default EP ratios; must be exactly Player.numEpRatios long.
	epRatios?: Array<number>;
	displayStats: Array<UnitStat>;
	// The Key view of the stats panel shows only these groups' weapon skills; unset shows none there.
	displayWeaponSkillGroups?: Array<WeaponSkillGroup>;
	modifyDisplayStats?: (player: Player<SpecType>) => StatMods;
	overwriteDisplayStats?: (player: Player<SpecType>) => StatWrites;

	// This can be used as a shorthand for setting "defaults".
	// Useful for when the defaults should be the same as the preset build options
	defaultBuild?: PresetBuild;
	defaults: {
		gear: EquipmentSpec;
		itemSwap?: ItemSwap;

		epWeights: Stats;
		// Used for Reforge Optimizer
		statCaps?: Stats;
		/**
		 * Allows specification of soft cap breakpoints for one or more stats.
		 *
		 * @remarks
		 * These function differently from the hard caps taken from the sim UI in a few ways:
		 *
		 * Firstly, the specified breakpoints are lower priority than hard caps, and
		 * evaluated only after the hard cap constraints have been solved first.
		 *
		 * Secondly, these constraints are evaluated in the order specified by the configuration
		 * Array rather than all at once. So once the hard caps have been respected, the
		 * closest breakpoint for the *first* listed soft capped stat is optimized against
		 * while ignoring any others. Then the solution is used to identify the closest
		 * breakpoint for the second listed stat (if present), etc.
		 */
		softCapBreakpoints?: StatCap[];
		breakpointLimits?: Stats;
		consumables: ConsumesSpec;
		talents: SavedTalents;
		specOptions: SpecOptions<SpecType>;

		raidBuffs: RaidBuffs;
		partyBuffs: PartyBuffs;
		individualBuffs: IndividualBuffs;

		debuffs: Debuffs;

		rotationType?: APLRotationType;
		// The APL a TypeAPL spec loads with. Without it the priority list starts empty.
		aplRotation?: APLRotation;
		simpleRotation?: SpecRotation<SpecType>;

		// Encounter applied by "Reset to Defaults" and on first load. Falls back to
		// the generic single-target dummy when unset.
		encounter?: PresetEncounter;

		other?: OtherDefaults;
	};

	playerInputs?: InputSection;
	playerIconInputs: Array<IconInputs.IconInputConfig<Player<SpecType>, any>>;
	petConsumeInputs?: Array<IconInputs.IconInputConfig<Player<SpecType>, any>>;
	rotationInputs?: InputSection;
	rotationIconInputs?: Array<IconInputs.IconInputConfig<Player<SpecType>, any>>;
	// Hides the Cooldowns block on the simple rotation tab. Set it where the spec's
	// `simpleRotation` generator ignores its `Cooldowns` argument, so the picker cannot
	// affect the generated APL and would only mislead.
	hideSimpleCooldowns?: boolean;
	includeBuffDebuffInputs: Array<any>;
	excludeBuffDebuffInputs: Array<any>;
	otherInputs: InputSection;
	// Currently, many classes don't support item swapping, and only in certain slots.
	// So enable it only where it is supported.
	itemSwapSlots?: Array<ItemSlot>;

	// Extra settings-tab sections, as data (e.g. Shaman totems).
	sections?: Array<CustomSection<SpecType>>;

	encounterPicker: EncounterPickerConfig;

	presets: {
		epWeights: Array<PresetEpWeights>;
		gear: Array<PresetGear>;
		talents: Array<SavedDataConfig<Player<SpecType>, SavedTalents>>;
		rotations: Array<PresetRotation>;
		encounters?: Array<PresetEncounter>;
		settings?: Array<PresetSettings>;
		builds?: Array<PresetBuild>;
		itemSwaps?: Array<PresetItemSwap>;
	};
}

// A setting whose value is derived from other settings (e.g. Monk stance from
// talents, DK AMS intake from the encounter). `apply` runs once at startup and
// again whenever `subscribe`'s source fires.
export interface DerivedSetting<SpecType extends Spec> {
	subscribe: (player: Player<SpecType>, sim: Sim) => StoreSubscribe;
	apply: (player: Player<SpecType>, sim: Sim) => void;
}

// The behaviour slots a spec declares on top of its config data. All optional:
// a spec that needs none of them declares none.
export interface SpecBehaviors<SpecType extends Spec> {
	// Wires the Reforge Optimizer. A function form receives the sim host, for
	// options that need to call back into it.
	reforge?: ReforgeOptimizerOptions | ((host: IndividualSimHost<SpecType>) => ReforgeOptimizerOptions);
	// Whether the player sims incoming healing. Defaults to `isTankSpec ||
	// isHealingSpec` from the spec's registry entry; set it only to override that.
	enableHealing?: boolean;
	derivedSettings?: Array<DerivedSetting<SpecType>>;
	// Spec-local escape hatch: anything else that needs constructing with the host.
	features?: Array<(host: IndividualSimHost<SpecType>) => unknown>;
}

// A whole spec as data. `ui/app/spec_entry.tsx` loads one of these per page.
export interface SpecDefinition<SpecType extends Spec> extends IndividualSimUIConfig<SpecType>, SpecBehaviors<SpecType> {
	spec: SpecType;
}

// Identity function; exists purely so `ui/specs/<class>/<spec>/spec.ts` gets the
// config checked against `SpecDefinition` without an `as` cast or a type
// annotation that would widen the literal spec type.
export function defineSpec<SpecType extends Spec>(def: SpecDefinition<SpecType>): SpecDefinition<SpecType> {
	return def;
}

// Removes dev-only gearsets from the gear presets and from the builds, dropping a
// build left empty. The default gear stays only when it is a production gearset
// the spec lists, so a default taken from an unlisted preset is cleared as well.
// `ui/app/spec_entry.tsx` applies it outside the dev server and the previews.
export function withoutDevOnlyGear<SpecType extends Spec>(def: SpecDefinition<SpecType>): SpecDefinition<SpecType> {
	const gear = def.presets.gear.filter(preset => !preset.devOnly);
	const builds = def.presets.builds
		?.map(build => (build.gear?.devOnly ? { ...build, gear: undefined } : build))
		.filter(({ name: _name, phase: _phase, group: _group, ...categories }) => Object.values(categories).some(value => value !== undefined));
	const productionGear = [...gear, ...(builds ?? []).map(build => build.gear)].map(preset => preset?.gear);
	return {
		...def,
		defaults: { ...def.defaults, gear: productionGear.includes(def.defaults.gear) ? def.defaults.gear : EquipmentSpec.create() },
		presets: { ...def.presets, gear, builds },
	};
}

export function registerSpecConfig<SpecType extends Spec>(spec: SpecType, config: IndividualSimUIConfig<SpecType>): IndividualSimUIConfig<SpecType> {
	registerPlayerConfig(spec, config);
	return config;
}

export const itemSwapEnabledSpecs: Array<any> = [];

export interface Settings {
	raidBuffs: RaidBuffs;
	partyBuffs: PartyBuffs;
	individualBuffs: IndividualBuffs;
	consumables: ConsumesSpec;
	race: Race;
	professions?: Array<Profession>;
}
