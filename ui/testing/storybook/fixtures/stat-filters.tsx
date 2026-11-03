import { STAT_GROUPS, StatFilterPopover } from '@features/stat-filters';
import { Class, PseudoStat, Stat } from '@generated/proto/common';
import { UnitStat } from '@sim/proto/stats';
import { fn } from 'storybook/test';

// Every character stats group, Hit (the rating stat) and Main Hand DPS for Gear, and Physical Damage, which no group holds and so lands under Other.
export const STAT_FILTER_UNIVERSE = [
	...STAT_GROUPS.flatMap(group => group.stats),
	UnitStat.fromStat(Stat.StatHitRating),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatMainHandDps),
	UnitStat.fromStat(Stat.StatPhysicalDamage),
];

export const WARRIOR_STAT_DEFAULTS = [
	UnitStat.fromStat(Stat.StatHealth),
	UnitStat.fromStat(Stat.StatStrength),
	UnitStat.fromStat(Stat.StatAgility),
	UnitStat.fromStat(Stat.StatAttackPower),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeHitPercent),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeCritPercent),
];

export const storyStatFilter = (testId: string, ariaLabel: string) => (
	<StatFilterPopover
		universe={STAT_FILTER_UNIVERSE}
		shown={WARRIOR_STAT_DEFAULTS}
		customised={false}
		playerClass={Class.ClassWarrior}
		onChange={fn()}
		onReset={fn()}
		groupsStorageKey={`storybook__${testId}Groups`}
		ariaLabel={ariaLabel}
		testId={testId}
	/>
);
