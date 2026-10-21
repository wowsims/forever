import { Class, Spec } from '@generated/proto/common';
import { LaunchStatus } from '@sim/constants/other';
import type { PlayerClass } from '@sim/player/player_class';

// The landing page's class order: alphabetical by class name.
export const LANDING_CLASS_ORDER: Class[] = [
	Class.ClassDruid,
	Class.ClassHunter,
	Class.ClassMage,
	Class.ClassPaladin,
	Class.ClassPriest,
	Class.ClassRogue,
	Class.ClassShaman,
	Class.ClassWarlock,
	Class.ClassWarrior,
];

// Each spec's theme class, as its spec.ts names it. A class card puts one on its cover to show that spec's
// background, which theme/specs.css sets as --theme-background-image.
export const SPEC_THEME_CLASS: Record<Exclude<Spec, Spec.SpecUnknown>, string> = {
	[Spec.SpecBalanceDruid]: 'balance-druid-sim-ui',
	[Spec.SpecFeralBearDruid]: 'feral-bear-druid-sim-ui',
	[Spec.SpecFeralCatDruid]: 'feral-druid-sim-ui',
	[Spec.SpecRestorationDruid]: 'restoration-druid-sim-ui',
	[Spec.SpecHunter]: 'hunter-sim-ui',
	[Spec.SpecMage]: 'mage-sim-ui',
	[Spec.SpecHolyPaladin]: 'holy-paladin-sim-ui',
	[Spec.SpecProtectionPaladin]: 'protection-paladin-sim-ui',
	[Spec.SpecRetributionPaladin]: 'retribution-paladin-sim-ui',
	[Spec.SpecDpsPriest]: 'priest-sim-ui',
	[Spec.SpecHealerPriest]: 'healer-priest-sim-ui',
	[Spec.SpecRogue]: 'rogue-sim-ui',
	[Spec.SpecElementalShaman]: 'elemental-shaman-sim-ui',
	[Spec.SpecEnhancementShaman]: 'enhancement-shaman-sim-ui',
	[Spec.SpecRestorationShaman]: 'restoration-shaman-sim-ui',
	[Spec.SpecWarlock]: 'warlock-sim-ui',
	[Spec.SpecDpsWarrior]: 'dps-warrior-sim-ui',
	[Spec.SpecProtectionWarrior]: 'protection-warrior-sim-ui',
};

// The furthest-along of the class's specs, so one unlaunched spec does not demote the class.
export const classLaunchStatus = (playerClass: PlayerClass<Class>): LaunchStatus =>
	Object.values(playerClass.specs).reduce<LaunchStatus>((best, spec) => Math.max(best, spec.launch.status), LaunchStatus.Unlaunched);
