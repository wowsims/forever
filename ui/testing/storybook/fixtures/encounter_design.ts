import {
	DamageAmpType,
	DesignBossAbility_Kind,
	DesignTargets,
	EncounterDesign,
	MobType,
	type PresetTarget,
	SpellSchool,
	Target,
} from '@generated/proto/common';

export const DESIGN_BASE_SECONDS = 180;

export const designTargetNames: ReadonlyArray<string> = ['Ragnaros', 'Son of Flame'];

export const designPresetTargets: ReadonlyArray<PresetTarget> = [
	{ path: 'Molten Core/Son of Flame', target: Target.create({ id: 12143, name: 'Son of Flame', level: 62, mobType: MobType.MobTypeElemental }) },
	{
		path: 'Blackwing Lair/Death Talon Wyrmguard',
		target: Target.create({ id: 12460, name: 'Death Talon Wyrmguard', level: 62, mobType: MobType.MobTypeDragonkin }),
	},
];

const at = (fraction: number, jitter = 0) => ({ at: fraction, jitter });
const noRepeat = { interval: 0, jitter: 0, count: 0 };

export const emptyDesign = (): EncounterDesign => EncounterDesign.create();

/** One wave of two adds every 45 s, alive for 20 s each time, with a debuff amp on the adds. */
export const dynamicAddsDesign = (): EncounterDesign =>
	EncounterDesign.create({
		addWaves: [
			{
				name: 'Sons of Flame',
				target: Target.clone(designPresetTargets[0].target!),
				count: 2,
				spawn: at(15 / 180, 2 / 180),
				lifetime: 20 / 180,
				repeat: { interval: 45 / 180, jitter: 3 / 180, count: 0 },
				executeProportion20: 0.2,
				executeProportion35: 0.35,
				killShare: 0.25,
			},
		],
		damageAmps: [
			{
				type: DamageAmpType.DamageAmpTypeTargetDebuff,
				percent: 10,
				start: at(15 / 180),
				end: at(35 / 180),
				repeat: { interval: 45 / 180, jitter: 0, count: 0 },
				targets: DesignTargets.create({ waveIndices: [0] }),
			},
		],
	});

/** A 10-yard move every 30 s, a few seconds either way. */
export const movementDesign = (): EncounterDesign =>
	EncounterDesign.create({
		moves: [
			{ at: at(20 / 180, 3 / 180), yards: 10, seconds: 0, repeat: { interval: 30 / 180, jitter: 0, count: 0 } },
			{ at: at(150 / 180), yards: 0, seconds: 4, repeat: noRepeat },
		],
	});

/** A tanking fight: a spell hit every 20 s, a strike, a thrash window and a swing pause, with health keyframes. */
export const customBossDesign = (): EncounterDesign =>
	EncounterDesign.create({
		bossHealth: [
			{ at: 0, healthPercent: 100 },
			{ at: 0.5, healthPercent: 40 },
			{ at: 1, healthPercent: 0 },
		],
		bossAbilities: [
			{
				kind: DesignBossAbility_Kind.SpellHit,
				targetIndex: 0,
				start: at(10 / 180),
				end: at(12 / 180),
				repeat: { interval: 20 / 180, jitter: 0, count: 0 },
				school: SpellSchool.SpellSchoolFire,
				minDamage: 1800,
				maxDamage: 2200,
				castTime: 2,
				weaponPercent: 0,
				procChance: 0,
				extraAttacks: 0,
				icd: 0,
				swingSpeedMultiplier: 1,
			},
			{
				kind: DesignBossAbility_Kind.Strike,
				targetIndex: 0,
				start: at(30 / 180),
				end: at(31 / 180),
				repeat: { interval: 30 / 180, jitter: 0, count: 3 },
				school: SpellSchool.SpellSchoolPhysical,
				minDamage: 0,
				maxDamage: 0,
				castTime: 0,
				weaponPercent: 150,
				procChance: 0,
				extraAttacks: 0,
				icd: 0,
				swingSpeedMultiplier: 1,
			},
			{
				kind: DesignBossAbility_Kind.Thrash,
				targetIndex: 0,
				start: at(60 / 180),
				end: at(100 / 180),
				repeat: noRepeat,
				school: SpellSchool.SpellSchoolPhysical,
				minDamage: 0,
				maxDamage: 0,
				castTime: 0,
				weaponPercent: 0,
				procChance: 20,
				extraAttacks: 2,
				icd: 2,
				swingSpeedMultiplier: 1,
			},
			{
				kind: DesignBossAbility_Kind.SwingTiming,
				targetIndex: 0,
				start: at(120 / 180, 2 / 180),
				end: at(128 / 180, 2 / 180),
				repeat: noRepeat,
				school: SpellSchool.SpellSchoolPhysical,
				minDamage: 0,
				maxDamage: 0,
				castTime: 0,
				weaponPercent: 0,
				procChance: 0,
				extraAttacks: 0,
				icd: 0,
				swingSpeedMultiplier: 0,
			},
		],
	});

/** Two downtimes on the boss that overlap, which validation reports as an error. */
export const overlappingDowntimesDesign = (): EncounterDesign =>
	EncounterDesign.create({
		downtimes: [
			{ targetIndex: 0, start: at(40 / 180), end: at(70 / 180), clearDebuffs: true },
			{ targetIndex: 0, start: at(60 / 180), end: at(90 / 180), clearDebuffs: false },
			{ targetIndex: 1, start: at(100 / 180), end: at(120 / 180), clearDebuffs: false },
		],
	});
